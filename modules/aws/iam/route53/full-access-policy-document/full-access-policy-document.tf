###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.route53_full_access.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "route53_full_access" {
  statement {
    effect = "Allow"

    actions = [
      "route53:*",
    ]

    resources = [
      "*",
    ]
  }
}
