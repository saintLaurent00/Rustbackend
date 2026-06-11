use sqlparser::dialect::GenericDialect;
use sqlparser::parser::Parser;
use sqlparser::ast::{Statement, Query, SetExpr, TableWithJoins, Expr, BinaryOperator};

pub fn inject_rls_conditions(sql: &str, rls_predicate: &str) -> String {
    let dialect = GenericDialect {};
    let ast = match Parser::parse_sql(&dialect, sql) {
        Ok(ast) => ast,
        Err(_) => return sql.to_string(),
    };

    let mut modified_ast = ast.clone();

    for statement in &mut modified_ast {
        if let Statement::Query(query) = statement {
            inject_into_query(query, rls_predicate);
        }
    }

    modified_ast.iter().map(|s| s.to_string()).collect::<Vec<_>>().join("; ")
}

fn inject_into_query(query: &mut Box<Query>, predicate: &str) {
    if let SetExpr::Select(select) = &mut *query.body {
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
            if let Some(existing_filter) = &mut select.selection {
                *existing_filter = Expr::BinaryOp {
                    left: Box::new(existing_filter.clone()),
                    op: BinaryOperator::And,
                    right: Box::new(new_filter),
                };
            } else {
                select.selection = Some(new_filter);
            }
        }
    }
}
