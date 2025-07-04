import React, { useState } from 'react'
import { Row, Col, Card, Statistic, Table, Tag, Space, Typography, Select } from 'antd'
import {
  LinkOutlined,
  EyeOutlined,
  GlobalOutlined,
  UserOutlined,
  ArrowUpOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { useSystemStats, useLinks } from '@/hooks/useApi'
import { useRealtimeStats } from '@/hooks/useStats'
import { StatsCharts } from '@/components/charts/StatsCharts'
import { formatNumber, getCountryFlag } from '@/utils/format'

const { Title } = Typography
const { Option } = Select

const Dashboard: React.FC = () => {
  const [timeRange, setTimeRange] = useState(24)
  const { data: systemStats, isLoading: statsLoading } = useSystemStats()
  const { data: linksData, isLoading: linksLoading } = useLinks(1, 10)
  const { data: realtimeStats, isLoading: realtimeLoading, refetch } = useRealtimeStats(timeRange)

  // Mock hourly data for the chart
  const hourlyData = Array.from({ length: 24 }, (_, i) => ({
    hour: `${i}:00`,
    hits: Math.floor(Math.random() * 1000) + 100,
  }))

  const pieColors = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8']

  const linkColumns = [
    {
      title: 'Link ID',
      dataIndex: 'link_id',
      key: 'link_id',
      render: (linkId: string) => (
        <code style={{ 
          background: '#f6f8fa', 
          padding: '2px 6px', 
          borderRadius: '4px',
          fontSize: '12px'
        }}>
          {linkId}
        </code>
      ),
    },
    {
      title: 'Business Unit',
      dataIndex: 'business_unit',
      key: 'business_unit',
      render: (bu: string) => <Tag color="blue">{bu}</Tag>,
    },
    {
      title: 'Network',
      dataIndex: 'network',
      key: 'network',
      render: (network: string) => <Tag color="green">{network}</Tag>,
    },
    {
      title: 'Hits',
      dataIndex: 'current_hits',
      key: 'current_hits',
      render: (hits: number) => formatNumber(hits),
      sorter: (a: any, b: any) => a.current_hits - b.current_hits,
    },
    {
      title: 'Cap',
      dataIndex: 'total_cap',
      key: 'total_cap',
      render: (cap: number) => cap > 0 ? formatNumber(cap) : 'Unlimited',
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'success' : 'default'}>
          {isActive ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          Dashboard
        </Title>
        <Space>
          <Select value={timeRange} onChange={setTimeRange} style={{ width: 120 }}>
            <Option value={6}>6 Hours</Option>
            <Option value={12}>12 Hours</Option>
            <Option value={24}>24 Hours</Option>
            <Option value={48}>48 Hours</Option>
            <Option value={168}>7 Days</Option>
          </Select>
          <Tag 
            icon={<ReloadOutlined spin={realtimeLoading} />} 
            color="processing" 
            onClick={() => refetch()}
            style={{ cursor: 'pointer' }}
          >
            {realtimeLoading ? 'Loading...' : 'Refresh'}
          </Tag>
        </Space>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Links"
              value={systemStats?.total_links || 0}
              prefix={<LinkOutlined />}
              loading={statsLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Hits"
              value={formatNumber(systemStats?.total_hits || 0)}
              prefix={<EyeOutlined />}
              loading={statsLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Today's Hits"
              value={formatNumber(systemStats?.today_hits || 0)}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#3f8600' }}
              loading={statsLoading}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Unique IPs"
              value={formatNumber(systemStats?.unique_ips || 0)}
              prefix={<UserOutlined />}
              loading={statsLoading}
            />
          </Card>
        </Col>
      </Row>

      {/* Real-time Statistics Charts */}
      {realtimeStats && <StatsCharts stats={realtimeStats} />}

      {/* Recent Links */}
      <Card 
        title="Recent Links" 
        extra={
          <Space>
            <Tag color="processing">Active Links</Tag>
          </Space>
        }
      >
        <Table
          columns={linkColumns}
          dataSource={linksData?.data || []}
          loading={linksLoading}
          pagination={false}
          size="middle"
          rowKey="id"
        />
      </Card>
    </div>
  )
}

export default Dashboard