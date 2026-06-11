use lettre::transport::smtp::authentication::Credentials;
use lettre::{Message, SmtpTransport, Transport};
use std::env;

pub async fn send_report(to: &str, subject: &str, body: &str) -> Result<(), Box<dyn std::error::Error>> {
    let smtp_host = env::var("SMTP_HOST").unwrap_or_else(|_| "localhost".to_string());
    let mail_from = env::var("MAIL_FROM").unwrap_or_else(|_| "noreply@bi-platform.local".to_string());

    let email = Message::builder()
        .from(mail_from.parse()?)
        .to(to.parse()?)
        .subject(subject)
        .body(body.to_string())?;

    let mailer = SmtpTransport::builder_relay(&smtp_host)?
        .build();

    mailer.send(&email)?;
    Ok(())
}
