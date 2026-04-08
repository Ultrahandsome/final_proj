package db

import (
	"CommentClassifier/internal/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

// GetDashboardData retrieves dashboard metrics including total comments, needs review count, pie chart, and bar chart data.
// It queries the MongoDB collection using aggregation pipelines for category counts and average confidence calculations.
// Returns a DashboardResponse containing computed data or an error if any query or operation fails.
func GetDashboardData(ctx context.Context) (*utils.DashboardResponse, error) {
	var response utils.DashboardResponse

	// Get total comments
	totalComments, err := CommentCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	response.TotalComments = totalComments

	// Get needs review
	needsReview, err := CommentCollection.CountDocuments(ctx,
		bson.M{"category": bson.M{"$in": []string{"Needs Review"}}},
	)
	if err != nil {
		return nil, err
	}
	response.NeedsReview = needsReview

	// Get Pie chart data: count per category
	var pieData []utils.PieChartData
	piePipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$category"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "category", Value: "$_id"},
			{Key: "count", Value: 1},
		}}},
	}
	pieCursor, err := CommentCollection.Aggregate(ctx, piePipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(pieCursor, ctx)

	if err := pieCursor.All(ctx, &pieData); err != nil {
		return nil, err
	}
	response.Pie = pieData

	// Get Bar chart data: average confidence score per category
	var barData []utils.BarChartData
	barPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$category"},
			{Key: "avgConf", Value: bson.D{{Key: "$avg", Value: "$confidenceScore"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "category", Value: "$_id"},
			{Key: "averageConfidence", Value: "$avgConf"},
		}}},
	}
	barCursor, err := CommentCollection.Aggregate(ctx, barPipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(barCursor, ctx)

	if err := barCursor.All(ctx, &barData); err != nil {
		return nil, err
	}
	response.Bar = barData

	return &response, nil
}
