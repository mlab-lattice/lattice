###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.assume_role_from_ec2_service.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "assume_role_from_ec2_service" {
  statement {
    effect = "Allow"

    actions = [
      "sts:AssumeRole",
    ]

    principals {
      type = "Service"

      identifiers = [
        "ec2.amazonaws.com",
      ]
    }
  }
}
