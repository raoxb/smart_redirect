import React, { useState } from 'react';
import { Row, Col, Card, Statistic, Typography, Tag, Button, Space, Tabs, Badge } from 'antd';
import {
  HeartOutlined,
  AlertOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertsList } from '@/components/monitoring/AlertsList';
import { api } from '@/services/api';

const { Title } = Typography;
const { TabPane } = Tabs;

interface HealthStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  timestamp: string;
  checks: {
    database: { status: string; latency: string };
    redis: { status: string; latency: string };
    api: { status: string; uptime: string };
  };
}

const Monitoring: React.FC = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('alerts');

  // Fetch active alerts
  const { data: alertsData, isLoading: alertsLoading, refetch: refetchAlerts } = useQuery({
    queryKey: ['monitoring', 'alerts'],
    queryFn: async () => {
      const response = await api.get('/monitor/alerts');
      return response.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Fetch health status
  const { data: healthData, isLoading: healthLoading, refetch: refetchHealth } = useQuery({
    queryKey: ['monitoring', 'health'],
    queryFn: async () => {
      const response = await api.get('/monitor/health');
      return response.data as HealthStatus;
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  });

  // Acknowledge alert mutation
  const acknowledgeMutation = useMutation({
    mutationFn: async (alertId: string) => {
      await api.post(`/monitor/alerts/${alertId}/acknowledge`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitoring', 'alerts'] });
    },
  });

  // Resolve alert mutation
  const resolveMutation = useMutation({
    mutationFn: async (alertId: string) => {
      await api.post(`/monitor/alerts/${alertId}/resolve`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitoring', 'alerts'] });
    },
  });

  const getHealthColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return '#52c41a';
      case 'degraded':
        return '#faad14';
      case 'unhealthy':
        return '#f5222d';
      default:
        return '#d9d9d9';
    }
  };

  const getHealthIcon = (status: string) => {
    return <HeartOutlined style={{ color: getHealthColor(status) }} />;
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          System Monitoring
        </Title>
        <Space>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={() => {
              refetchAlerts();
              refetchHealth();
            }}
          >
            Refresh
          </Button>
        </Space>
      </div>

      {/* Health Status Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="System Status"
              value={healthData?.status || 'Unknown'}
              valueStyle={{ color: getHealthColor(healthData?.status || '') }}
              prefix={getHealthIcon(healthData?.status || '')}
              loading={healthLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Database"
              value={healthData?.checks.database.status || 'Unknown'}
              suffix={
                <Tag color={healthData?.checks.database.status === 'healthy' ? 'green' : 'red'}>
                  {healthData?.checks.database.latency || 'N/A'}
                </Tag>
              }
              prefix={<DatabaseOutlined />}
              loading={healthLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Redis"
              value={healthData?.checks.redis.status || 'Unknown'}
              suffix={
                <Tag color={healthData?.checks.redis.status === 'healthy' ? 'green' : 'red'}>
                  {healthData?.checks.redis.latency || 'N/A'}
                </Tag>
              }
              prefix={<CloudServerOutlined />}
              loading={healthLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Active Alerts"
              value={alertsData?.count || 0}
              valueStyle={{ color: alertsData?.count > 0 ? '#f5222d' : '#52c41a' }}
              prefix={<AlertOutlined />}
              loading={alertsLoading}
            />
          </Card>
        </Col>
      </Row>

      {/* Tabs for different monitoring views */}
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane 
            tab={
              <span>
                <AlertOutlined />
                Active Alerts
                {alertsData?.count > 0 && (
                  <Badge count={alertsData.count} style={{ marginLeft: 8 }} />
                )}
              </span>
            } 
            key="alerts"
          >
            <AlertsList
              alerts={alertsData?.alerts || []}
              loading={alertsLoading}
              onAcknowledge={(id) => acknowledgeMutation.mutate(id)}
              onResolve={(id) => resolveMutation.mutate(id)}
            />
          </TabPane>
          
          <TabPane tab="Alert History" key="history">
            <div style={{ padding: '20px', textAlign: 'center', color: '#999' }}>
              Alert history view coming soon...
            </div>
          </TabPane>
          
          <TabPane tab="Configuration" key="config">
            <div style={{ padding: '20px', textAlign: 'center', color: '#999' }}>
              Monitoring configuration coming soon...
            </div>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

export default Monitoring;