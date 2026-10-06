import importlib.util
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("terraform-plan-guard.py")
SPEC = importlib.util.spec_from_file_location("terraform_plan_guard", MODULE_PATH)
assert SPEC and SPEC.loader
GUARD = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(GUARD)


def change(address: str, resource_type: str, actions: list[str]) -> dict:
    return {
        "address": address,
        "type": resource_type,
        "change": {"actions": actions},
    }


def planned_lambda(address: str, image_uri: str) -> dict:
    return {
        "address": address,
        "type": "aws_lambda_function",
        "values": {"image_uri": image_uri},
    }


class TerraformPlanGuardTests(unittest.TestCase):
    def test_allows_create_and_update(self) -> None:
        summary = GUARD.inspect_plan(
            {
                "resource_changes": [
                    change("aws_iam_role.producer", "aws_iam_role", ["create"]),
                    change(
                        "aws_lambda_function.api[0]",
                        "aws_lambda_function",
                        ["update"],
                    ),
                    change("aws_s3_bucket.web", "aws_s3_bucket", ["no-op"]),
                ]
            }
        )

        self.assertEqual(
            summary["counts"],
            {"create": 1, "update": 1, "delete": 0, "replace": 0},
        )
        self.assertEqual(summary["blocked"], [])

    def test_blocks_delete_and_both_replace_orders(self) -> None:
        summary = GUARD.inspect_plan(
            {
                "resource_changes": [
                    change(
                        "aws_dynamodb_table.main",
                        "aws_dynamodb_table",
                        ["delete"],
                    ),
                    change(
                        "aws_s3_bucket.raw",
                        "aws_s3_bucket",
                        ["delete", "create"],
                    ),
                    change(
                        "aws_cloudfront_distribution.web",
                        "aws_cloudfront_distribution",
                        ["create", "delete"],
                    ),
                ]
            }
        )

        self.assertEqual(
            summary["counts"],
            {"create": 0, "update": 0, "delete": 1, "replace": 2},
        )
        self.assertEqual(len(summary["blocked"]), 3)
        self.assertTrue(all("protected resource" in item for item in summary["blocked"]))

    def test_blocks_unknown_actions(self) -> None:
        summary = GUARD.inspect_plan(
            {"resource_changes": [change("example.unknown", "example", ["forget"])]}
        )

        self.assertEqual(len(summary["blocked"]), 1)
        self.assertIn("unsupported actions", summary["blocked"][0])

    def test_requires_both_lambdas_to_use_candidate_digest(self) -> None:
        image_uri = "example.dkr.ecr.eu-west-1.amazonaws.com/cloudops@sha256:abc"
        summary = GUARD.inspect_plan(
            {
                "planned_values": {
                    "root_module": {
                        "resources": [
                            planned_lambda("aws_lambda_function.api[0]", image_uri),
                            planned_lambda(
                                "aws_lambda_function.worker[0]", image_uri
                            ),
                        ]
                    }
                }
            },
            expected_image_uri=image_uri,
        )

        self.assertEqual(summary["blocked"], [])

    def test_blocks_mismatched_candidate_digest(self) -> None:
        expected = "example.dkr.ecr.eu-west-1.amazonaws.com/cloudops@sha256:abc"
        summary = GUARD.inspect_plan(
            {
                "planned_values": {
                    "root_module": {
                        "resources": [
                            planned_lambda(
                                "aws_lambda_function.api[0]",
                                "example.dkr.ecr.eu-west-1.amazonaws.com/cloudops@sha256:def",
                            ),
                            planned_lambda(
                                "aws_lambda_function.worker[0]", expected
                            ),
                        ]
                    }
                }
            },
            expected_image_uri=expected,
        )

        self.assertEqual(len(summary["blocked"]), 1)
        self.assertIn("does not match candidate", summary["blocked"][0])

    def test_blocks_missing_planned_lambda(self) -> None:
        expected = "example.dkr.ecr.eu-west-1.amazonaws.com/cloudops@sha256:abc"
        summary = GUARD.inspect_plan(
            {"planned_values": {"root_module": {"resources": []}}},
            expected_image_uri=expected,
        )

        self.assertEqual(len(summary["blocked"]), 2)
        self.assertTrue(all("missing from planned values" in item for item in summary["blocked"]))


if __name__ == "__main__":
    unittest.main()
