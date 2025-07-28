import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Select,
  Switch,
  message,
  Tabs,
  Table,
  Space,
  Modal,
  Tooltip,
} from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { analysisService } from '../services/analysis';
import { authService } from '../services/auth';
import { LLMConfig, LLMConfigRequest, User } from '../services/types';

const { Option } = Select;
const { TabPane } = Tabs;
const { confirm } = Modal;

const Settings: React.FC = () => {
  const [llmConfigs, setLlmConfigs] = useState<LLMConfig[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<LLMConfig | null>(null);
  const [form] = Form.useForm();
  const [profileForm] = Form.useForm();

  const loadLLMConfigs = useCallback(async () => {
    try {
      const data = await analysisService.getLLMConfig();
      setLlmConfigs(data);
    } catch (error: any) {
      message.error('加载LLM配置失败: ' + error.message);
    }
  }, []);

  const loadUserProfile = useCallback(async () => {
    try {
      const userData = await authService.getProfile();
      setUser(userData);
      profileForm.setFieldsValue(userData);
    } catch (error: any) {
      message.error('加载用户信息失败: ' + error.message);
    }
  }, [profileForm]);

  useEffect(() => {
    loadLLMConfigs();
    loadUserProfile();
  }, [loadLLMConfigs, loadUserProfile]);

  const handleAddLLMConfig = () => {
    setEditingConfig(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEditLLMConfig = (config: LLMConfig) => {
    setEditingConfig(config);
    form.setFieldsValue(config);
    setModalVisible(true);
  };

  const handleDeleteLLMConfig = (config: LLMConfig) => {
    confirm({
      title: '确认删除',
      content: `确定要删除 ${config.provider} (${config.model}) 的配置吗？`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          // 这里应该有删除API，暂时从本地状态删除
          setLlmConfigs(prev => prev.filter(item => item.id !== config.id));
          message.success('删除成功！');
        } catch (error: any) {
          message.error('删除失败: ' + error.message);
        }
      },
    });
  };

  const handleSubmitLLMConfig = async (values: LLMConfigRequest) => {
    setLoading(true);
    try {
      const newConfig = await analysisService.configLLM(values);
      if (editingConfig) {
        setLlmConfigs(prev => prev.map(item => 
          item.id === editingConfig.id ? newConfig : item
        ));
        message.success('更新配置成功！');
      } else {
        setLlmConfigs(prev => [newConfig, ...prev]);
        message.success('添加配置成功！');
      }
      setModalVisible(false);
    } catch (error: any) {
      message.error('保存配置失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateProfile = async (values: Partial<User>) => {
    setLoading(true);
    try {
      const updatedUser = await authService.updateProfile(values);
      setUser(updatedUser);
      message.success('更新个人资料成功！');
    } catch (error: any) {
      message.error('更新失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const llmColumns: ColumnsType<LLMConfig> = [
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      render: (provider) => {
        const providerMap: Record<string, { name: string; color: string }> = {
          openai: { name: 'OpenAI', color: '#10a37f' },
          hunyuan: { name: '腾讯混元', color: '#1890ff' },
          tongyi: { name: '通义千问', color: '#722ed1' },
        };
        const config = providerMap[provider] || { name: provider, color: '#666' };
        return <span style={{ color: config.color, fontWeight: 'bold' }}>{config.name}</span>;
      },
    },
    {
      title: '模型',
      dataIndex: 'model',
      key: 'model',
    },
    {
      title: 'API密钥',
      dataIndex: 'api_key',
      key: 'api_key',
      render: (key) => `${key.substring(0, 8)}...${key.substring(key.length - 8)}`,
    },
    {
      title: '默认',
      dataIndex: 'is_default',
      key: 'is_default',
      render: (isDefault) => <Switch checked={isDefault} disabled />,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEditLLMConfig(record)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteLLMConfig(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <h1>设置</h1>
      
      <Tabs defaultActiveKey="llm">
        <TabPane tab="LLM配置" key="llm">
          <Card
            title="大语言模型配置"
            extra={
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleAddLLMConfig}
              >
                添加配置
              </Button>
            }
          >
            <Table
              columns={llmColumns}
              dataSource={llmConfigs}
              rowKey="id"
              pagination={false}
            />
          </Card>
        </TabPane>

        <TabPane tab="个人资料" key="profile">
          <Card title="个人资料">
            <Form
              form={profileForm}
              layout="vertical"
              onFinish={handleUpdateProfile}
              style={{ maxWidth: 400 }}
            >
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input />
              </Form.Item>

              <Form.Item
                label="邮箱"
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' }
                ]}
              >
                <Input />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading}>
                  更新资料
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>
      </Tabs>

      {/* LLM配置模态框 */}
      <Modal
        title={editingConfig ? '编辑LLM配置' : '添加LLM配置'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmitLLMConfig}
        >
          <Form.Item
            label="提供商"
            name="provider"
            rules={[{ required: true, message: '请选择提供商' }]}
          >
            <Select placeholder="选择LLM提供商">
              <Option value="openai">OpenAI</Option>
              <Option value="hunyuan">腾讯混元</Option>
              <Option value="tongyi">通义千问</Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="模型"
            name="model"
            rules={[{ required: true, message: '请输入模型名称' }]}
          >
            <Input placeholder="如: gpt-3.5-turbo, hunyuan-lite" />
          </Form.Item>

          <Form.Item
            label="API密钥"
            name="api_key"
            rules={[{ required: true, message: '请输入API密钥' }]}
          >
            <Input.Password
              placeholder="输入API密钥"
              iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            />
          </Form.Item>

          <Form.Item
            label="设为默认"
            name="is_default"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {editingConfig ? '更新' : '添加'}
              </Button>
              <Button onClick={() => setModalVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings;
