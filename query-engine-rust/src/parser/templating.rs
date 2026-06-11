use tera::{Tera, Context};
use serde_json::Value;

pub fn compile_dynamic_sql(raw_sql: &str, filters: &Value) -> String {
    let mut tera = Tera::default();
    let mut context = Context::new();

    if let Some(obj) = filters.as_object() {
        for (k, v) in obj {
            context.insert(k, v);
        }
    }

    match tera.render_str(raw_sql, &context) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Error rendering SQL: {}", e);
            raw_sql.to_string()
        }
    }
}
