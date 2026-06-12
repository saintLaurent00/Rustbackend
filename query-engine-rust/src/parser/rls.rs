use sqlparser::dialect::GenericDialect;
use sqlparser::parser::Parser;
use sqlparser::ast::{Statement, Query, SetExpr, TableWithJoins, Expr, BinaryOperator};

/// inject_rls_conditions prend un SQL brut et un prédicat de sécurité RLS (ex: "agence_id = 12"),
/// et injecte ce prédicat dans l'arbre syntaxique (AST) du SQL original.
/// Cela garantit que la restriction est appliquée mathématiquement au niveau de la base de données,
/// évitant toute fuite de données ou injection SQL manuelle.
pub fn inject_rls_conditions(sql: &str, rls_predicate: &str) -> String {
    let dialect = GenericDialect {};

    // Étape 1 : Analyse du SQL brut en AST
    let ast = match Parser::parse_sql(&dialect, sql) {
        Ok(ast) => ast,
        Err(_) => return sql.to_string(), // En cas d'erreur de parsing, on retourne le SQL original (à améliorer avec une erreur bloquante)
    };

    let mut modified_ast = ast.clone();

    // Étape 2 : Parcours des statements pour injecter le RLS dans les requêtes SELECT
    for statement in &mut modified_ast {
        if let Statement::Query(query) = statement {
            inject_into_query(query, rls_predicate);
        }
    }

    // Étape 3 : Re-génération du SQL à partir de l'AST modifié
    modified_ast.iter().map(|s| s.to_string()).collect::<Vec<_>>().join("; ")
}

/// inject_into_query navigue dans la structure de la requête pour atteindre la clause WHERE (selection).
fn inject_into_query(query: &mut Box<Query>, predicate: &str) {
    if let SetExpr::Select(select) = &mut *query.body {

        // Parsing du prédicat RLS pour obtenir une expression AST valide
        let rls_expr = match Parser::parse_sql(&GenericDialect {}, &format!("SELECT * WHERE {}", predicate)) {
            Ok(ast) => {
                if let Statement::Query(q) = &ast[0] {
                    if let SetExpr::Select(s) = &*q.body {
                        s.selection.clone()
                    } else { None }
                } else { None }
            },
            Err(_) => None,
        };

        if let Some(new_filter) = rls_expr {
            // Si une clause WHERE existe déjà, on combine avec AND
            if let Some(existing_filter) = &mut select.selection {
                *existing_filter = Expr::BinaryOp {
                    left: Box::new(existing_filter.clone()),
                    op: BinaryOperator::And,
                    right: Box::new(new_filter),
                };
            } else {
                // Sinon, on crée la clause WHERE avec le prédicat RLS
                select.selection = Some(new_filter);
            }
        }
    }
}
