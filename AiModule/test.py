#!/usr/bin/env python3
import unittest
import grpc
import comment_pb2
import comment_pb2_grpc
import random
import string
from collections.abc import Iterable

class TestCommentClassifierService(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.channel = grpc.insecure_channel('localhost:50051')
        cls.stub = comment_pb2_grpc.CommentClassifierStub(cls.channel)

    @classmethod
    def tearDownClass(cls):
        cls.channel.close()

    def send_request(self, raw_comments):
        """Helper to send request."""
        request = comment_pb2.RawComments(comments=raw_comments)
        responses = list(self.stub.ClassifyComments(request))
        classified_comments = []
        for resp in responses:
            classified_comments.extend(resp.comments)
        return classified_comments

    def validate_classified_comment(self, comment):
        """Validate structure and contents of a ClassifiedComment."""
        # Basic type checks
        self.assertIsInstance(comment.id, str, "id should be a string")
        self.assertIsInstance(comment.label, str, "label should be a string")
        self.assertIsInstance(comment.similarComment, Iterable, "similarComment should be iterable")
        self.assertIsInstance(comment.keywords, Iterable, "keywords should be iterable")
        self.assertIsInstance(comment.score, float, "score should be a float")

        # ID should not be empty
        self.assertTrue(comment.id.strip(), "id should not be empty")

        # Label should be within allowed labels
        allowed_labels = {
            "Safe", "Needs Review", "Harassment", "Swearing",
            "Hate Speech", "Sarcasm", "Aggressive", "Complaint",
            "Constructive Feedback", "Other", "Error"
        }
        self.assertIn(comment.label, allowed_labels, f"label {comment.label} not in allowed labels")

        # Score should be between 0.0 and 1.0
        self.assertGreaterEqual(comment.score, 0.0, "score should be >= 0.0")
        self.assertLessEqual(comment.score, 1.0, "score should be <= 1.0")

        # Each item in similarComment should be non-empty string
        for idx, sim_comment in enumerate(comment.similarComment):
            self.assertIsInstance(sim_comment, str, f"similarComment[{idx}] should be a string")
            self.assertTrue(sim_comment.strip(), f"similarComment[{idx}] should not be empty")

        # Each keyword should be a non-empty string
        for idx, keyword in enumerate(comment.keywords):
            self.assertIsInstance(keyword, str, f"keyword[{idx}] should be a string")
            self.assertTrue(keyword.strip(), f"keyword[{idx}] should not be empty")

    def test_normal_comments(self):
        """Test normal comments."""
        raw_comments = [
            comment_pb2.RawComment(id="test1", rawComment="This is a safe test comment."),
            comment_pb2.RawComment(id="test2", rawComment="You are stupid! This course sucks!"),
        ]
        classified_comments = self.send_request(raw_comments)
        self.assertEqual(len(classified_comments), 2)

        for c in classified_comments:
            with self.subTest(comment_id=c.id):
                self.validate_classified_comment(c)

    def test_empty_input(self):
        """Test sending an empty comment list."""
        raw_comments = []
        classified_comments = self.send_request(raw_comments)
        self.assertEqual(len(classified_comments), 0)

    def test_long_comment(self):
        """Test sending a very long comment."""
        long_text = "This is a very long comment. " * 500
        raw_comments = [
            comment_pb2.RawComment(id="long1", rawComment=long_text)
        ]
        classified_comments = self.send_request(raw_comments)
        self.assertEqual(len(classified_comments), 1)
        self.validate_classified_comment(classified_comments[0])

    def test_special_characters(self):
        """Test special character comment."""
        special_text = "@#$%^&*()_+-=[]{}|;':,./<>?"
        raw_comments = [
            comment_pb2.RawComment(id="special1", rawComment=special_text)
        ]
        classified_comments = self.send_request(raw_comments)
        self.assertEqual(len(classified_comments), 1)
        self.validate_classified_comment(classified_comments[0])

    def test_random_batch(self):
        """Test batch of random short comments."""
        raw_comments = []
        for i in range(20):
            text = ''.join(random.choices(string.ascii_letters + " ", k=random.randint(10, 100)))
            raw_comments.append(
                comment_pb2.RawComment(id=f"rand{i}", rawComment=text)
            )
        classified_comments = self.send_request(raw_comments)
        self.assertEqual(len(classified_comments), 20)

        for c in classified_comments:
            with self.subTest(comment_id=c.id):
                self.validate_classified_comment(c)

if __name__ == '__main__':
    unittest.main()
