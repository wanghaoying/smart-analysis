import React from 'react';
import { Card, Alert } from 'antd';
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { VisualizationResponse } from '../services/types';

interface ChartDisplayProps {
  data: VisualizationResponse;
}

interface ChartData {
  name: string;
  value: number;
}

interface ChartConfig {
  data: {
    labels?: string[];
    datasets?: Array<{ data: number[] }>;
  } | ChartData[];
}

const ChartDisplay: React.FC<ChartDisplayProps> = ({ data }) => {
  const colors = ['#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#00ff00', '#ff00ff'];

  // 尝试解析图表数据
  let chartConfig: ChartConfig;
  try {
    chartConfig = typeof data.chart_data === 'string' 
      ? JSON.parse(data.chart_data) 
      : data.chart_data;
  } catch (error) {
    console.error('解析图表数据失败:', error);
    return (
      <Alert
        message="图表数据格式错误"
        description="无法解析返回的图表数据，请检查数据格式。"
        type="error"
        showIcon
      />
    );
  }

  if (!chartConfig || !chartConfig.data) {
    return (
      <Alert
        message="无图表数据"
        description="AI返回的数据中没有包含有效的图表配置。"
        type="warning"
        showIcon
      />
    );
  }

  const renderChart = () => {
    // 处理图表数据格式转换
    let chartData: ChartData[];
    
    if (Array.isArray(chartConfig.data)) {
      // 如果 data 已经是 ChartData[] 格式
      chartData = chartConfig.data as ChartData[];
    } else {
      // 如果 data 是包含 labels 和 datasets 的对象格式
      const configData = chartConfig.data as { labels?: string[]; datasets?: Array<{ data: number[] }> };
      if (configData.datasets && configData.labels) {
        chartData = configData.labels.map((label: string, index: number) => ({
          name: label,
          value: configData.datasets![0].data[index],
        }));
      } else {
        chartData = [];
      }
    }

    switch (data.chart_type) {
      case 'bar':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="value" fill="#8884d8" />
            </BarChart>
          </ResponsiveContainer>
        );

      case 'line':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="value" stroke="#8884d8" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        );

      case 'pie':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <PieChart>
              <Pie
                data={chartData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }: any) => `${name} ${((percent || 0) * 100).toFixed(0)}%`}
                outerRadius={120}
                fill="#8884d8"
                dataKey="value"
              >
                {chartData.map((entry: any, index: number) => (
                  <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        );

      default:
        return (
          <Alert
            message="不支持的图表类型"
            description={`图表类型 "${data.chart_type}" 暂不支持显示。`}
            type="info"
            showIcon
          />
        );
    }
  };

  return (
    <Card title={data.title || '数据可视化图表'}>
      {renderChart()}
      
      {/* 显示原始配置（调试用） */}
      {process.env.NODE_ENV === 'development' && (
        <details style={{ marginTop: 16 }}>
          <summary style={{ cursor: 'pointer', color: '#666' }}>
            查看原始数据 (开发模式)
          </summary>
          <pre style={{ 
            fontSize: 12, 
            backgroundColor: '#f5f5f5', 
            padding: 8, 
            borderRadius: 4,
            marginTop: 8,
            overflow: 'auto'
          }}>
            {JSON.stringify(chartConfig, null, 2)}
          </pre>
        </details>
      )}
    </Card>
  );
};

export default ChartDisplay;
