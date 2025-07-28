import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Card,
  Select,
  Button,
  Input,
  List,
  Spin,
  message,
  Modal,
  Space,
  Empty,
  Divider,
} from 'antd';
import {
  SendOutlined,
  FileTextOutlined,
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import { fileService } from '../services/file';
import { analysisService } from '../services/analysis';
import { FileInfo, Session, Query, QueryRequest, VisualizationRequest } from '../services/types';
import ChatMessage from '../components/ChatMessage';
import ChartDisplay from '../components/ChartDisplay';

const { Option } = Select;
const { TextArea } = Input;

const Analysis: React.FC = () => {
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [queries, setQueries] = useState<Query[]>([]);
  const [selectedFile, setSelectedFile] = useState<FileInfo | null>(null);
  const [question, setQuestion] = useState('');
  const [loading, setLoading] = useState(false);
  const [sessionLoading, setSessionLoading] = useState(false);
  const [chartModalVisible, setChartModalVisible] = useState(false);
  const [chartData, setChartData] = useState<any>(null);

  const loadFiles = async () => {
    try {
      const data = await fileService.getList();
      setFiles(data.filter(file => file.status === 'ready'));
    } catch (error: any) {
      message.error('加载文件列表失败: ' + error.message);
    }
  };

  const loadSessions = async () => {
    // 这里应该有获取会话列表的API，暂时用空数组
    setSessions([]);
  };

  const loadQueries = useCallback(async () => {
    if (!currentSession) return;
    
    try {
      const data = await analysisService.getHistory(currentSession.id);
      setQueries(data || []);
    } catch (error: any) {
      message.error('加载对话历史失败: ' + error.message);
      setQueries([]);
    }
  }, [currentSession]);

  useEffect(() => {
    loadFiles();
    loadSessions();
  }, []);

  useEffect(() => {
    if (currentSession) {
      loadQueries();
    }
  }, [currentSession, loadQueries]);

  const handleCreateSession = async () => {
    if (!selectedFile) {
      message.warning('请先选择一个文件');
      return;
    }

    setSessionLoading(true);
    try {
      const session = await analysisService.createSession(
        `${selectedFile.orig_name} 分析会话`,
        selectedFile.id
      );
      setCurrentSession(session);
      setSessions(prev => [session, ...(prev || [])]);
      setQueries([]);
      message.success('创建会话成功！');
    } catch (error: any) {
      message.error('创建会话失败: ' + error.message);
    } finally {
      setSessionLoading(false);
    }
  };

  const handleSendQuery = async () => {
    if (!question.trim()) {
      message.warning('请输入问题');
      return;
    }

    if (!currentSession) {
      message.warning('请先创建会话');
      return;
    }

    const queryRequest: QueryRequest = {
      session_id: currentSession.id,
      question: question.trim(),
      file_id: selectedFile?.id,
    };

    setLoading(true);
    try {
      const response = await analysisService.query(queryRequest);
      
      // 创建新的查询记录
      const newQuery: Query = {
        id: Date.now(), // 临时ID
        session_id: currentSession.id,
        user_id: 0, // 暂时用0
        question: question.trim(),
        answer: response.answer,
        query_type: response.query_type,
        status: response.status,
        created_at: new Date().toISOString(),
      };

      setQueries(prev => [...(prev || []), newQuery]);
      setQuestion('');
      message.success('查询完成！');
    } catch (error: any) {
      message.error('查询失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateChart = async (chartType: string) => {
    if (!currentSession || !selectedFile) {
      message.warning('请先选择文件并创建会话');
      return;
    }

    if (!question.trim()) {
      message.warning('请输入要可视化的问题');
      return;
    }

    const request: VisualizationRequest = {
      session_id: currentSession.id,
      query: question.trim(),
      file_id: selectedFile.id,
      chart_type: chartType,
    };

    setLoading(true);
    try {
      const response = await analysisService.visualize(request);
      setChartData(response);
      setChartModalVisible(true);
      message.success('图表生成成功！');
    } catch (error: any) {
      message.error('生成图表失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const quickQuestions = [
    '这个数据集包含多少行和多少列？',
    '数据的基本统计信息是什么？',
    '有哪些缺失值吗？',
    '各个列的数据类型是什么？',
    '显示前10行数据',
  ];

  return (
    <div>
      <h1>数据分析</h1>
      
      <Row gutter={[16, 16]}>
        {/* 左侧控制面板 */}
        <Col xs={24} lg={8}>
          <Card title="分析设置" size="small">
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', marginBottom: 8 }}>选择文件：</label>
              <Select
                style={{ width: '100%' }}
                placeholder="请选择要分析的文件"
                value={selectedFile?.id}
                onChange={(value) => {
                  const file = files.find(f => f.id === value);
                  setSelectedFile(file || null);
                }}
              >
                {files.map(file => (
                  <Option key={file.id} value={file.id}>
                    <Space>
                      <FileTextOutlined />
                      {file.orig_name}
                    </Space>
                  </Option>
                ))}
              </Select>
            </div>

            <Button
              type="primary"
              block
              onClick={handleCreateSession}
              loading={sessionLoading}
              disabled={!selectedFile}
            >
              开始分析
            </Button>

            {currentSession && (
              <div style={{ marginTop: 16, padding: 12, background: '#f5f5f5', borderRadius: 6 }}>
                <div style={{ fontSize: 12, color: '#666' }}>当前会话</div>
                <div style={{ fontWeight: 'bold' }}>{currentSession.name}</div>
              </div>
            )}
          </Card>

          <Card title="快速问题" size="small" style={{ marginTop: 16 }}>
            <List
              size="small"
              dataSource={quickQuestions}
              renderItem={item => (
                <List.Item
                  style={{ cursor: 'pointer', padding: '8px 0' }}
                  onClick={() => setQuestion(item)}
                >
                  <div style={{ fontSize: 12, color: '#666' }}>{item}</div>
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* 右侧对话区域 */}
        <Col xs={24} lg={16}>
          <Card 
            title="AI 对话" 
            size="small"
            extra={
              <Space>
                <Button
                  icon={<BarChartOutlined />}
                  size="small"
                  onClick={() => handleGenerateChart('bar')}
                  disabled={!currentSession || loading}
                >
                  柱状图
                </Button>
                <Button
                  icon={<LineChartOutlined />}
                  size="small"
                  onClick={() => handleGenerateChart('line')}
                  disabled={!currentSession || loading}
                >
                  折线图
                </Button>
                <Button
                  icon={<PieChartOutlined />}
                  size="small"
                  onClick={() => handleGenerateChart('pie')}
                  disabled={!currentSession || loading}
                >
                  饼图
                </Button>
              </Space>
            }
          >
            {/* 对话历史 */}
            <div style={{ height: 400, overflowY: 'auto', marginBottom: 16 }}>
              {(queries || []).length === 0 ? (
                <Empty 
                  description="暂无对话记录"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  style={{ marginTop: 100 }}
                />
              ) : (
                <List
                  dataSource={queries || []}
                  renderItem={item => (
                    <div style={{ marginBottom: 16 }}>
                      <ChatMessage
                        type="user"
                        content={item.question}
                        timestamp={item.created_at}
                      />
                      <ChatMessage
                        type="assistant"
                        content={item.answer}
                        timestamp={item.created_at}
                      />
                    </div>
                  )}
                />
              )}
              {loading && (
                <div style={{ textAlign: 'center', marginTop: 20 }}>
                  <Spin size="small" />
                  <div style={{ marginTop: 8, color: '#666' }}>AI正在思考中...</div>
                </div>
              )}
            </div>

            <Divider style={{ margin: '16px 0' }} />

            {/* 输入区域 */}
            <div>
              <TextArea
                rows={3}
                placeholder="输入你的问题，比如：这个数据集的销售趋势如何？"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                onPressEnter={(e) => {
                  if (e.ctrlKey || e.metaKey) {
                    handleSendQuery();
                  }
                }}
              />
              <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ fontSize: 12, color: '#999' }}>
                  按 Ctrl/Cmd + Enter 发送
                </div>
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSendQuery}
                  loading={loading}
                  disabled={!currentSession}
                >
                  发送
                </Button>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 图表显示模态框 */}
      <Modal
        title="数据可视化"
        open={chartModalVisible}
        onCancel={() => setChartModalVisible(false)}
        width={800}
        footer={null}
      >
        {chartData && <ChartDisplay data={chartData} />}
      </Modal>
    </div>
  );
};

export default Analysis;
