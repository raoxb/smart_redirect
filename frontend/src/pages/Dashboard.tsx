import React from 'react'
import { Row, Col, Card, Statistic, Table, Tag, Space, Typography } from 'antd'
import {
  LinkOutlined,
  EyeOutlined,
  GlobalOutlined,
  UserOutlined,
  ArrowUpOutlined,
} from '@ant-design/icons'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import { useSystemStats, useLinks } from '@/hooks/useApi'
import { formatNumber, getCountryFlag } from '@/utils/format'

const { Title } = Typography

const Dashboard: React.FC = () => {
  const { data: systemStats, isLoading: statsLoading } = useSystemStats()
  const { data: linksData, isLoading: linksLoading } = useLinks(1, 10)

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
      <Title level={2} style={{ marginBottom: 24 }}>
        Dashboard
      </Title>

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

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {/* Traffic Chart */}
        <Col xs={24} lg={16}>
          <Card title="24-Hour Traffic" extra={<Tag color="blue">Last 24 Hours</Tag>}>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={hourlyData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="hour" />
                <YAxis />
                <Tooltip />
                <Line 
                  type="monotone" 
                  dataKey="hits" 
                  stroke="#1890ff" 
                  strokeWidth={2}
                  dot={{ fill: '#1890ff', strokeWidth: 2, r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>

        {/* Top Countries */}
        <Col xs={24} lg={8}>
          <Card title="Traffic by Country" extra={<GlobalOutlined />}>
            {systemStats?.top_countries ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={systemStats.top_countries.slice(0, 5)}
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="hits"
                    label={({ country, hits }) => `${getCountryFlag(country)} ${formatNumber(hits)}`}
                  >
                    {systemStats.top_countries.slice(0, 5).map((_, index) => (
                      <Cell key={`cell-${index}`} fill={pieColors[index % pieColors.length]} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(value) => formatNumber(value as number)} />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div style={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                No data available
              </div>
            )}
          </Card>
        </Col>
      </Row>

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