import React from 'react';
import { List, Tag, Button, Card, Space, Typography, Empty, Spin } from 'antd';
import { 
  ExclamationCircleOutlined, 
  WarningOutlined, 
  InfoCircleOutlined,
  CheckCircleOutlined
} from '@ant-design/icons';
import { formatDistanceToNow } from 'date-fns';

const { Text } = Typography;

interface Alert {
  id: string;
  type: string;
  level: 'info' | 'warning' | 'critical';
  title: string;
  message: string;
  details: Record<string, any>;
  created_at: string;
  resolved_at?: string;
  acknowledged: boolean;
}

interface AlertsListProps {
  alerts: Alert[];
  loading?: boolean;
  onAcknowledge?: (alertId: string) => void;
  onResolve?: (alertId: string) => void;
}

const AlertIcon: React.FC<{ level: string }> = ({ level }) => {
  switch (level) {
    case 'critical':
      return <ExclamationCircleOutlined style={{ color: '#f5222d' }} />;
    case 'warning':
      return <WarningOutlined style={{ color: '#faad14' }} />;
    case 'info':
      return <InfoCircleOutlined style={{ color: '#1890ff' }} />;
    default:
      return <InfoCircleOutlined />;
  }
};

const AlertLevel: React.FC<{ level: string }> = ({ level }) => {
  const colors = {
    critical: 'error',
    warning: 'warning',
    info: 'processing',
  };
  
  return (
    <Tag color={colors[level as keyof typeof colors] || 'default'}>
      {level.toUpperCase()}
    </Tag>
  );
};

export const AlertsList: React.FC<AlertsListProps> = ({
  alerts,
  loading = false,
  onAcknowledge,
  onResolve,
}) => {
  if (loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px' }}>
          <Spin size="large" />
        </div>
      </Card>
    );
  }

  if (!alerts.length) {
    return (
      <Card>
        <Empty 
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="No active alerts"
        />
      </Card>
    );
  }

  return (
    <List
      dataSource={alerts}
      renderItem={(alert) => (
        <Card
          style={{ marginBottom: 16 }}
          bodyStyle={{ padding: '16px' }}
        >
          <List.Item
            actions={[
              !alert.acknowledged && onAcknowledge && (
                <Button 
                  size="small" 
                  onClick={() => onAcknowledge(alert.id)}
                >
                  Acknowledge
                </Button>
              ),
              !alert.resolved_at && onResolve && (
                <Button 
                  size="small" 
                  type="primary"
                  onClick={() => onResolve(alert.id)}
                  icon={<CheckCircleOutlined />}
                >
                  Resolve
                </Button>
              ),
            ].filter(Boolean)}
          >
            <List.Item.Meta
              avatar={<AlertIcon level={alert.level} />}
              title={
                <Space>
                  <Text strong>{alert.title}</Text>
                  <AlertLevel level={alert.level} />
                  {alert.acknowledged && <Tag color="blue">Acknowledged</Tag>}
                  {alert.resolved_at && <Tag color="green">Resolved</Tag>}
                </Space>
              }
              description={
                <>
                  <Text>{alert.message}</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {formatDistanceToNow(new Date(alert.created_at), { addSuffix: true })}
                  </Text>
                  {alert.details && Object.keys(alert.details).length > 0 && (
                    <div style={{ marginTop: 8 }}>
                      {Object.entries(alert.details).map(([key, value]) => (
                        <Tag key={key} style={{ marginBottom: 4 }}>
                          {key}: {typeof value === 'object' ? JSON.stringify(value) : value}
                        </Tag>
                      ))}
                    </div>
                  )}
                </>
              }
            />
          </List.Item>
        </Card>
      )}
    />
  );
};