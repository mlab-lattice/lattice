###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.s3_full_access.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "s3_full_access" {
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      "*",
    ]
  }
}
