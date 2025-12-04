resource "aws_ses_domain_identity" "this" {
  domain = data.aws_route53_zone.this.name
}

resource "aws_route53_record" "verification_record" {
  zone_id = data.aws_route53_zone.this.zone_id
  name    = "_amazonses.${data.aws_route53_zone.this.name}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.this.verification_token]
}

resource "aws_ses_domain_identity_verification" "verification" {
  domain = aws_ses_domain_identity.this.domain

  depends_on = [aws_route53_record.verification_record]
}

resource "aws_sesv2_configuration_set" "this" {
  configuration_set_name = "${var.prefix}-volleygoals"
}

resource "aws_route53_record" "dmarc" {
  zone_id = data.aws_route53_zone.this.zone_id
  name    = "_dmarc.${data.aws_route53_zone.this.name}"
  type    = "TXT"
  ttl     = "600"
  records = ["v=DMARC1; p=none; rua=mailto:dmarc-reports@${data.aws_route53_zone.this.name}"]
}

resource "aws_route53_record" "mail" {
  zone_id = data.aws_route53_zone.this.zone_id
  name    = data.aws_route53_zone.this.name
  type    = "MX"
  ttl     = "600"
  records = ["10 inbound-smtp.${data.aws_region.current.region}.amazonaws.com"]
}

resource "aws_route53_record" "spf" {
  zone_id = data.aws_route53_zone.this.zone_id
  name    = data.aws_route53_zone.this.name
  type    = "TXT"
  ttl     = "600"
  records = ["v=spf1 include:amazonses.com ~all"]
}

resource "aws_ses_template" "invitation" {
  name    = "${var.prefix}-invitation"
  subject = "You're invited to join {{teamName}} on VolleyGoals"
  html    = <<-HTML
    <!doctype html>
    <html>
    <head>
      <meta charset="utf-8" />
      <meta name="viewport" content="width=device-width,initial-scale=1" />
      <style>
        /* Base (light) theme derived from your MUI theme */
        body{font-family:Arial,Helvetica,sans-serif;background:#ffffff;color:#000000;margin:0;padding:0}
        .email-container{max-width:600px;margin:24px auto;background:#f8f8f8;border-radius:8px;overflow:hidden;box-shadow:0 2px 6px rgba(0,0,0,.06)}
        .header{padding:24px;background:#C41E3A;color:#ffffff;text-align:center}
        .content{padding:24px;color:#000000}
        a{color:#C41E3A}
        .button{display:inline-block;padding:12px 20px;background:#C41E3A;color:#ffffff;text-decoration:none;border-radius:6px}
        .footer{padding:16px;font-size:12px;color:#666666;text-align:center}

        /* Dark-mode hint: some email clients support prefers-color-scheme */
        @media (prefers-color-scheme: dark) {
          body{background:#0a0a0a;color:#ffffff}
          .email-container{background:#1a1a1a;box-shadow:none}
          .header{background:#C41E3A;color:#ffffff}
          .content{color:#ffffff}
          .button{background:#C41E3A;color:#ffffff}
          .footer{color:#b0b0b0}
          a{color:#C41E3A}
        }
      </style>
    </head>
    <body>
      <div class="email-container">
        <div class="header">
          <h1 style="margin:0;font-size:20px">You're invited to join {{teamName}}</h1>
        </div>
        <div class="content">
          <p>Hello and welcome to VolleyGoals!</p>
          <p><strong>{{inviterName}}</strong> invited you to join the <strong>{{teamName}}</strong> team on VolleyGoals so you can share goals and collaborate.</p>
          <p style="text-align:center">
            <a class="button" href="{{acceptLink}}" target="_blank" rel="noopener">Accept Invitation</a>
          </p>
          <p style="margin-top:16px"><strong>Message from {{inviterName}}:</strong></p>
          <p style="white-space:pre-wrap;margin-top:6px;color:#374151">{{personalMessage}}</p>
          <p>If the button doesn't work, copy and paste this link into your browser:</p>
          <p><a href="{{acceptLink}}" target="_blank" rel="noopener">{{acceptLink}}</a></p>
          <p style="color:#6b7280">This invitation will expire in {{expiryDays}} days.</p>
        </div>
        <div class="footer">This message was sent from <strong>no-reply@${data.aws_route53_zone.this.name}</strong>. Please do not reply to this email. For help or support, visit <a href="https://${data.aws_route53_zone.this.name}/support" target="_blank" rel="noopener">VolleyGoals Support</a>.</div>
      </div>
    </body>
    </html>
  HTML
  text    = <<-TEXT
    Hello and welcome to VolleyGoals!

    {{inviterName}} invited you to join the "{{teamName}}" team on VolleyGoals to share goals and collaborate.

    Accept invitation: {{acceptLink}}

    Message from {{inviterName}}:

    {{personalMessage}}

    This invitation will expire in {{expiryDays}} days.

    This message was sent from no-reply@${data.aws_route53_zone.this.name}. Please do not reply to this email.
    For support visit: https://${data.aws_route53_zone.this.name}/support

    Thanks,
    The VolleyGoals team
  TEXT
}
