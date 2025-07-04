import React, { useState } from 'react';
import {
  Row,
  Col,
  Card,
  Statistic,
  Typography,
  Tag,
  Button,
  Space,
  Tabs,
  Badge,
  Form,
  InputNumber,
  Switch,
  Select,
  Table,
  DatePicker,
  Timeline,
} from 'antd';
import {
  HeartOutlined,
  AlertOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  ReloadOutlined,
  SettingOutlined,
  HistoryOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { AlertsList } from '@/components/monitoring/AlertsList';
import { api } from '@/services/api';
import { Line } from '@ant-design/plots';

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

interface MonitoringConfig {
  error_rate_threshold: number;
  response_time_threshold: number;
  traffic_spike_threshold: number;
  check_interval: number;
  alert_cooldown: number;
  enable_email_alerts: boolean;
  enable_webhook_alerts: boolean;
}

interface AlertHistory {
  id: string;
  type: string;
  severity: string;
  message: string;
  created_at: string;
  resolved_at?: string;
  acknowledged_at?: string;
}

const { RangePicker } = DatePicker;
const { Option } = Select;

const Monitoring: React.FC = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('alerts');
  const [configForm] = Form.useForm();

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

  // Fetch monitoring configuration
  const { data: configData, isLoading: configLoading } = useQuery({
    queryKey: ['monitoring', 'config'],
    queryFn: async () => {
      const response = await api.get('/monitor/config');
      return response.data as MonitoringConfig;
    },
  });

  // Fetch alert history
  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ['monitoring', 'history'],
    queryFn: async () => {
      // For now, return mock data since the API might not be implemented
      return {
        data: [
          {
            id: '1',
            type: 'error_rate',
            severity: 'warning',
            message: 'Error rate exceeded threshold: 5.2%',
            created_at: '2025-07-04T17:30:00Z',
            resolved_at: '2025-07-04T17:45:00Z',
          },
          {
            id: '2',
            type: 'response_time',
            severity: 'critical',
            message: 'Response time exceeded 2000ms',
            created_at: '2025-07-04T16:15:00Z',
            acknowledged_at: '2025-07-04T16:20:00Z',
            resolved_at: '2025-07-04T16:30:00Z',
          },
        ] as AlertHistory[]
      };
    },
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

  // Update config mutation
  const updateConfigMutation = useMutation({
    mutationFn: async (config: MonitoringConfig) => {
      await api.put('/monitor/config', config);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitoring', 'config'] });
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
          
          <TabPane 
            tab={
              <span>
                <HistoryOutlined />
                Alert History
              </span>
            } 
            key="history"
          >
            <AlertHistoryView 
              history={historyData?.data || []} 
              loading={historyLoading} 
            />
          </TabPane>
          
          <TabPane 
            tab={
              <span>
                <SettingOutlined />
                Configuration
              </span>
            } 
            key="config"
          >
            <MonitoringConfigView
              config={configData}
              loading={configLoading}
              onUpdate={(config) => updateConfigMutation.mutate(config)}
              form={configForm}
            />
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
};

// Alert History Component
const AlertHistoryView: React.FC<{
  history: AlertHistory[];
  loading: boolean;
}> = ({ history, loading }) => {
  const columns = [
    {
      title: 'Time',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
      sorter: (a: AlertHistory, b: AlertHistory) => 
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color="blue">{type.replace('_', ' ').toUpperCase()}</Tag>
      ),
    },
    {
      title: 'Severity',
      dataIndex: 'severity',
      key: 'severity',
      render: (severity: string) => {
        const color = severity === 'critical' ? 'red' : 
                     severity === 'warning' ? 'orange' : 'green';
        return <Tag color={color}>{severity.toUpperCase()}</Tag>;
      },
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
    },
    {
      title: 'Status',
      key: 'status',
      render: (record: AlertHistory) => {
        if (record.resolved_at) {
          return <Tag color="green">Resolved</Tag>;
        } else if (record.acknowledged_at) {
          return <Tag color="orange">Acknowledged</Tag>;
        } else {
          return <Tag color="red">Active</Tag>;
        }
      },
    },
    {
      title: 'Duration',
      key: 'duration',
      render: (record: AlertHistory) => {
        const start = new Date(record.created_at);
        const end = record.resolved_at ? new Date(record.resolved_at) : new Date();
        const duration = Math.round((end.getTime() - start.getTime()) / 1000 / 60);
        return `${duration}m`;
      },
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: '16px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <h3>Alert History</h3>
          </Col>
          <Col>
            <RangePicker />
          </Col>
        </Row>
      </div>
      
      <Table
        columns={columns}
        dataSource={history}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />

      {history.length > 0 && (
        <Card title="Alert Timeline" style={{ marginTop: '24px' }}>
          <Timeline>
            {history.slice(0, 5).map((alert) => (
              <Timeline.Item
                key={alert.id}
                color={alert.severity === 'critical' ? 'red' : 
                       alert.severity === 'warning' ? 'orange' : 'blue'}
              >
                <div>
                  <strong>{alert.type.replace('_', ' ').toUpperCase()}</strong>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    {new Date(alert.created_at).toLocaleString()}
                  </div>
                  <div>{alert.message}</div>
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        </Card>
      )}
    </div>
  );
};

// Monitoring Configuration Component
const MonitoringConfigView: React.FC<{
  config?: MonitoringConfig;
  loading: boolean;
  onUpdate: (config: MonitoringConfig) => void;
  form: any;
}> = ({ config, loading, onUpdate, form }) => {
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      onUpdate(values);
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  // Set form values when config is loaded
  React.useEffect(() => {
    if (config) {
      form.setFieldsValue(config);
    }
  }, [config, form]);

  return (
    <div>
      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card title="Alert Thresholds" loading={loading}>
            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              initialValues={{
                error_rate_threshold: 5,
                response_time_threshold: 2000,
                traffic_spike_threshold: 200,
                check_interval: 60,
                alert_cooldown: 300,
                enable_email_alerts: false,
                enable_webhook_alerts: false,
              }}
            >
              <Form.Item
                name="error_rate_threshold"
                label="Error Rate Threshold (%)"
                rules={[{ required: true, message: 'Please enter error rate threshold' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  max={100}
                  step={0.1}
                  placeholder="5.0"
                />
              </Form.Item>

              <Form.Item
                name="response_time_threshold"
                label="Response Time Threshold (ms)"
                rules={[{ required: true, message: 'Please enter response time threshold' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={100}
                  placeholder="2000"
                />
              </Form.Item>

              <Form.Item
                name="traffic_spike_threshold"
                label="Traffic Spike Threshold (%)"
                rules={[{ required: true, message: 'Please enter traffic spike threshold' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={10}
                  placeholder="200"
                />
              </Form.Item>

              <Form.Item
                name="check_interval"
                label="Check Interval (seconds)"
                rules={[{ required: true, message: 'Please enter check interval' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={10}
                  step={10}
                  placeholder="60"
                />
              </Form.Item>

              <Form.Item
                name="alert_cooldown"
                label="Alert Cooldown (seconds)"
                rules={[{ required: true, message: 'Please enter alert cooldown' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={60}
                  step={60}
                  placeholder="300"
                />
              </Form.Item>

              <Button type="primary" htmlType="submit" block>
                Update Configuration
              </Button>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="Notification Settings" loading={loading}>
            <Form form={form} layout="vertical">
              <Form.Item
                name="enable_email_alerts"
                label="Email Alerts"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name="enable_webhook_alerts"
                label="Webhook Alerts"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Form>

            <div style={{ marginTop: '24px' }}>
              <h4>Current Settings</h4>
              {config && (
                <div style={{ fontSize: '12px', color: '#666' }}>
                  <p>Error Rate: {config.error_rate_threshold}%</p>
                  <p>Response Time: {config.response_time_threshold}ms</p>
                  <p>Traffic Spike: {config.traffic_spike_threshold}%</p>
                  <p>Check Interval: {config.check_interval}s</p>
                  <p>Alert Cooldown: {config.alert_cooldown}s</p>
                  <p>Email Alerts: {config.enable_email_alerts ? 'Enabled' : 'Disabled'}</p>
                  <p>Webhook Alerts: {config.enable_webhook_alerts ? 'Enabled' : 'Disabled'}</p>
                </div>
              )}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Monitoring;