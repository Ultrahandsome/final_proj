import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Spin, Alert } from 'antd';
import { FileTextOutlined, WarningOutlined } from '@ant-design/icons';
import { request } from 'umi';
import { Pie, Column } from '@ant-design/plots';

// Type definitions for API response
interface PieChartData {
  category: string;
  count: number;
}

interface BarChartData {
  category: string;
  averageConfidence: number;
}

interface DashboardData {
  totalComments: number;
  needsReview: number;
  pie: PieChartData[];
  bar: BarChartData[];
}

interface ApiResponse {
  code: number;
  data: DashboardData;
  msg: string;
}

const DashboardCharts: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [dashboardData, setDashboardData] = useState<DashboardData>({
    totalComments: 0,
    needsReview: 0,
    pie: [],
    bar: [],
  });

  const fetchDashboardData = async (): Promise<void> => {
    try {
      setLoading(true);
      const response = await request<ApiResponse>('/api/dashboard');

      // Check if API returns the expected structure with code, data, and msg
      if (response.code === 200 && response.data) {
        setDashboardData(response.data);
        setError(null);
      } else {
        setError(`API Error: ${response.msg || 'Unknown error'}`);
        console.error('API returned an error:', response);
      }
    } catch (err) {
      setError('Failed to fetch dashboard data. Please try again later.');
      console.error('Error fetching dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, []);

  // Pie chart configuration
  const pieConfig = {
    appendPadding: 10,
    data: dashboardData.pie || [],
    angleField: 'count',
    colorField: 'category',
    label: {
      text: 'category',
      style: {
        fontWeight: 'bold',
      },
    },
    legend: {
      color: {
        title: false,
        position: 'right',
        rowPadding: 5,
      },
    },
  };

  // Bar chart configuration
  const barConfig = {
    data: dashboardData.bar || [],
    xField: 'category',
    yField: 'averageConfidence',
    label: {
      text: (d: any) => `${(d.averageConfidence * 100).toFixed(1)}%`,
      textBaseline: 'bottom',
    },
    axis: {
      y: {
        labelFormatter: '.00%',
      },
    },
  };
  const pieData = dashboardData.pie ?? [];
  const barData = dashboardData.bar ?? [];
  return (
    <>
      {error && (
        <Alert
          message="Error"
          description={error}
          type="error"
          showIcon
          style={{ marginBottom: 24 }}
        />
      )}

      <Spin spinning={loading}>
        {/* Key Metrics */}
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12}>
            <Card>
              <Statistic
                title="Total Comments"
                value={dashboardData.totalComments}
                prefix={<FileTextOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12}>
            <Card>
              <Statistic
                title="Needs Review"
                value={dashboardData.needsReview}
                prefix={<WarningOutlined />}
                valueStyle={{ color: dashboardData.needsReview > 0 ? '#faad14' : '#3f8600' }}
              />
            </Card>
          </Col>
        </Row>

        {/* Charts */}
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col xs={24} lg={12}>
            <Card title="Comment Categories Distribution">
              {pieData.length > 0 ? (
                <Pie {...pieConfig} />
              ) : (
                <div style={{ textAlign: 'center', padding: 24 }}>No data available</div>
              )}
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card title="Average Confidence by Category">
              {barData.length > 0 ? (
                <Column {...barConfig} />
              ) : (
                <div style={{ textAlign: 'center', padding: 24 }}>No data available</div>
              )}
            </Card>
          </Col>
        </Row>
      </Spin>
    </>
  );
};

export default DashboardCharts;
