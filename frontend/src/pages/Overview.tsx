import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Progress, List, Avatar } from 'antd';
import {
  FileTextOutlined,
  BarChartOutlined,
  CloudOutlined,
  TrophyOutlined,
} from '@ant-design/icons';
import { fileService } from '../services/file';
import { analysisService } from '../services/analysis';
import { FileInfo, Query, UsageResponse } from '../services/types';

const Overview: React.FC = () => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [recentQueries, setRecentQueries] = useState<Query[]>([]);
  const [usage, setUsage] = useState<UsageResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [filesData, queriesData, usageData] = await Promise.all([
        fileService.getList().catch(() => []),
        analysisService.getHistory().catch(() => []),
        analysisService.getUsage().catch(() => null),
      ]);
      
      setFiles(filesData || []);
      setRecentQueries((queriesData || []).slice(0, 5)); // 最近5条查询
      setUsage(usageData);
    } catch (error) {
      console.error('加载数据失败:', error);
      setFiles([]);
      setRecentQueries([]);
      setUsage(null);
    } finally {
      setLoading(false);
    }
  };

  const getFileStatusCount = (status: string) => {
    return (files || []).filter(file => file.status === status).length;
  };

  return (
    <div>
      <h1>概览</h1>
      
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总文件数"
              value={files?.length || 0}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="已处理文件"
              value={getFileStatusCount('ready')}
              prefix={<BarChartOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总查询次数"
              value={recentQueries?.length || 0}
              prefix={<TrophyOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="消耗Token"
              value={usage?.total_tokens || 0}
              prefix={<CloudOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        {/* 文件状态分布 */}
        <Col xs={24} lg={12}>
          <Card title="文件状态分布" loading={loading}>
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>已就绪</span>
                <span>{getFileStatusCount('ready')}/{files?.length || 0}</span>
              </div>
              <Progress 
                percent={files?.length ? (getFileStatusCount('ready') / files.length) * 100 : 0} 
                status="success"
              />
            </div>
            
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>处理中</span>
                <span>{getFileStatusCount('processing')}/{files?.length || 0}</span>
              </div>
              <Progress 
                percent={files?.length ? (getFileStatusCount('processing') / files.length) * 100 : 0}
                status="active"
              />
            </div>
            
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>错误</span>
                <span>{getFileStatusCount('error')}/{files?.length || 0}</span>
              </div>
              <Progress 
                percent={files?.length ? (getFileStatusCount('error') / files.length) * 100 : 0}
                status="exception"
              />
            </div>
          </Card>
        </Col>

        {/* 最近查询 */}
        <Col xs={24} lg={12}>
          <Card title="最近查询" loading={loading}>
            <List
              itemLayout="horizontal"
              dataSource={recentQueries || []}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<Avatar icon={<BarChartOutlined />} />}
                    title={item.question?.length > 50 ? item.question.substring(0, 50) + '...' : item.question}
                    description={`${item.query_type} • ${new Date(item.created_at).toLocaleString()}`}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      {/* 使用量统计 */}
      {usage && (
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col span={24}>
            <Card title="使用量统计">
              <Row gutter={16}>
                <Col xs={24} sm={8}>
                  <Statistic
                    title="总消耗Token"
                    value={usage.total_tokens}
                    suffix="tokens"
                  />
                </Col>
                <Col xs={24} sm={8}>
                  <Statistic
                    title="预估费用"
                    value={usage.total_cost}
                    precision={4}
                    prefix="$"
                  />
                </Col>
                <Col xs={24} sm={8}>
                  <Statistic
                    title="平均每次查询"
                    value={(recentQueries?.length || 0) ? Math.round(usage.total_tokens / recentQueries.length) : 0}
                    suffix="tokens"
                  />
                </Col>
              </Row>
            </Card>
          </Col>
        </Row>
      )}
    </div>
  );
};

export default Overview;
