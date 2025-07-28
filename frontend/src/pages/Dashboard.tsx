import React, { useState, useEffect } from 'react';
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Button,
  message,
} from 'antd';
import {
  DashboardOutlined,
  FileTextOutlined,
  BarChartOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { useNavigate, Outlet, useLocation } from 'react-router-dom';
import { authService } from '../services/auth';
import { User } from '../services/types';
import './Dashboard.css';

const { Header, Sider, Content } = Layout;

const Dashboard: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    // 检查用户是否已登录
    if (!authService.isAuthenticated()) {
      navigate('/login');
      return;
    }

    // 获取用户信息
    const currentUser = authService.getCurrentUser();
    setUser(currentUser);
  }, [navigate]);

  const handleMenuClick = (key: string) => {
    navigate(`/dashboard/${key}`);
  };

  const handleLogout = () => {
    authService.logout();
    message.success('退出登录成功');
  };

  const userMenu = (
    <Menu>
      <Menu.Item key="profile" icon={<UserOutlined />}>
        个人资料
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key="logout" icon={<LogoutOutlined />} onClick={handleLogout}>
        退出登录
      </Menu.Item>
    </Menu>
  );

  const sidebarItems = [
    {
      key: 'overview',
      icon: <DashboardOutlined />,
      label: '概览',
    },
    {
      key: 'files',
      icon: <FileTextOutlined />,
      label: '文件管理',
    },
    {
      key: 'analysis',
      icon: <BarChartOutlined />,
      label: '数据分析',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
  ];

  // 获取当前选中的菜单项
  const getSelectedKey = () => {
    const path = location.pathname.split('/').pop();
    return path === 'dashboard' ? 'overview' : path || 'overview';
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        theme="light"
        width={220}
      >
        <div className="logo">
          <h2>AI分析平台</h2>
        </div>
        <Menu
          theme="light"
          selectedKeys={[getSelectedKey()]}
          mode="inline"
          items={sidebarItems}
          onClick={({ key }) => handleMenuClick(key)}
        />
      </Sider>
      
      <Layout>
        <Header className="header">
          <div className="header-left">
            <Button
              type="text"
              onClick={() => setCollapsed(!collapsed)}
              style={{
                fontSize: '16px',
                width: 64,
                height: 64,
              }}
            />
          </div>
          
          <div className="header-right">
            <Dropdown overlay={userMenu} placement="bottomRight">
              <div className="user-info">
                <Avatar icon={<UserOutlined />} />
                <span className="username">{user?.username}</span>
              </div>
            </Dropdown>
          </div>
        </Header>
        
        <Content className="content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default Dashboard;
