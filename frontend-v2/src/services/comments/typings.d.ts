declare namespace Comment {
  type Comment = {
    _id: string;
    rawComment: string;
    category: string;
    confidenceScore: number;
    keywords: string[];
    lastUpdated: number;
    isDeleted: boolean;
    updateHistory: UpdateHistory[];
    similarComments: string[];
  };

  type UpdateHistory = {
    time: number;
    user: string;
    category: string;
    comment: string;
  };

  type GetCommentsRequest = {
    page: number;
    limit: number;
    categories: string[];
  };

  // Dashboard data type definitions
  type PieChartData = {
    category: string;
    count: number;
  };

  type BarChartData = {
    category: string;
    averageConfidence: number;
  };

  type DashboardData = {
    totalComments: number;
    needsReview: number;
    pie: PieChartData[];
    bar: BarChartData[];
  };

  type ApiResponse<T> = {
    code: number;
    data: T;
    msg: string;
  };

  type DashboardResponse = ApiResponse<DashboardData>;
}
