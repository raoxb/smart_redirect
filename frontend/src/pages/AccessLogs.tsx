import React, { useState } from 'react'
import { 
  Table, 
  Card, 
  Input, 
  Select, 
  DatePicker, 
  Tag, 
  Space, 
  Button,
  Typography,
  Tooltip,
  Row,
  Col
} from 'antd'
import { SearchOutlined, ReloadOutlined, EyeOutlined, LinkOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/services/api'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { Title } = Typography
const { Option } = Select
const { RangePicker } = DatePicker

interface AccessLog {
  id: number
  link_id: number
  target_id: number
  ip: string
  user_agent: string
  referer: string
  country: string
  client_ip: string
  created_at: string
  target?: {
    id: number
    url: string
    weight: number
    countries: string
  }
  link?: {
    id: number
    link_id: string
    business_unit: string
  }
}

interface AccessLogsResponse {
  data: AccessLog[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

const AccessLogs: React.FC = () => {
  const [filters, setFilters] = useState({
    page: 1,
    page_size: 20,
    link_id: '',
    ip: '',
    country: ''
  })

  // Fetch access logs
  const { data: logsData, isLoading, refetch } = useQuery<AccessLogsResponse>({
    queryKey: ['accessLogs', filters],
    queryFn: () => {
      const params = new URLSearchParams()
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.append(key, value.toString())
      })
      return api.get(`/stats/access-logs?${params.toString()}`).then(res => res.data)
    },
  })

  // Fetch links for filter dropdown
  const { data: linksData } = useQuery({
    queryKey: ['links'],
    queryFn: () => api.get('/links').then(res => res.data),
  })

  const handleTableChange = (pagination: any) => {
    setFilters(prev => ({
      ...prev,
      page: pagination.current,
      page_size: pagination.pageSize
    }))
  }

  const handleFilterChange = (key: string, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1 // Reset to first page when filtering
    }))
  }

  const clearFilters = () => {
    setFilters({
      page: 1,
      page_size: 20,
      link_id: '',
      ip: '',
      country: ''
    })
  }

  const getCountryFlag = (country: string) => {
    const flags: { [key: string]: string } = {
      'US': '🇺🇸', 'CN': '🇨🇳', 'GB': '🇬🇧', 'DE': '🇩🇪', 
      'FR': '🇫🇷', 'IT': '🇮🇹', 'AU': '🇦🇺', 'TW': '🇹🇼',
      'LOCAL': '🏠'
    }
    return flags[country] || '🌍'
  }

  const columns: ColumnsType<AccessLog> = [
    {
      title: 'Time',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => (
        <Tooltip title={dayjs(date).format('YYYY-MM-DD HH:mm:ss')}>
          {dayjs(date).format('MM-DD HH:mm:ss')}
        </Tooltip>
      ),
      sorter: true,
    },
    {
      title: 'Link',
      key: 'link',
      width: 120,
      render: (record: AccessLog) => {
        if (record.link) {
          return (
            <Space direction="vertical" size={0}>
              <Tag color="blue" icon={<LinkOutlined />}>
                {record.link.link_id}
              </Tag>
              <span style={{ fontSize: '12px', color: '#666' }}>
                {record.link.business_unit}
              </span>
            </Space>
          )
        }
        return <Tag>Unknown</Tag>
      },
    },
    {
      title: 'Target URL',
      key: 'target',
      width: 200,
      render: (record: AccessLog) => {
        if (record.target) {
          const url = new URL(record.target.url)
          return (
            <Tooltip title={record.target.url}>
              <Space direction="vertical" size={0}>
                <span style={{ fontWeight: 500 }}>
                  {url.hostname}
                </span>
                <span style={{ fontSize: '12px', color: '#666' }}>
                  {url.pathname}
                </span>
              </Space>
            </Tooltip>
          )
        }
        return <span style={{ color: '#999' }}>No target</span>
      },
    },
    {
      title: 'IP Address',
      dataIndex: 'ip',
      key: 'ip',
      width: 140,
      render: (ip: string) => (
        <Tag color="purple">{ip}</Tag>
      ),
      filterDropdown: ({ setSelectedKeys, selectedKeys, confirm, clearFilters }) => (
        <div style={{ padding: 8 }}>
          <Input
            placeholder="Search IP"
            value={selectedKeys[0]}
            onChange={e => setSelectedKeys(e.target.value ? [e.target.value] : [])}
            onPressEnter={() => {
              confirm()
              handleFilterChange('ip', selectedKeys[0])
            }}
            style={{ width: 188, marginBottom: 8, display: 'block' }}
          />
          <Space>
            <Button
              type="primary"
              onClick={() => {
                confirm()
                handleFilterChange('ip', selectedKeys[0])
              }}
              icon={<SearchOutlined />}
              size="small"
              style={{ width: 90 }}
            >
              Search
            </Button>
            <Button
              onClick={() => {
                clearFilters?.()
                handleFilterChange('ip', '')
              }}
              size="small"
              style={{ width: 90 }}
            >
              Reset
            </Button>
          </Space>
        </div>
      ),
      filterIcon: (filtered: boolean) => (
        <SearchOutlined style={{ color: filtered ? '#1890ff' : undefined }} />
      ),
    },
    {
      title: 'Country',
      dataIndex: 'country',
      key: 'country',
      width: 100,
      render: (country: string) => (
        <Space>
          <span>{getCountryFlag(country)}</span>
          <span>{country}</span>
        </Space>
      ),
    },
    {
      title: 'User Agent',
      dataIndex: 'user_agent',
      key: 'user_agent',
      width: 250,
      render: (userAgent: string) => {
        const getBrowser = (ua: string) => {
          if (ua.includes('Chrome')) return 'Chrome'
          if (ua.includes('Firefox')) return 'Firefox'
          if (ua.includes('Safari')) return 'Safari'
          if (ua.includes('Edge')) return 'Edge'
          return 'Unknown'
        }
        
        const getOS = (ua: string) => {
          if (ua.includes('Windows')) return 'Windows'
          if (ua.includes('Mac')) return 'macOS'
          if (ua.includes('Linux')) return 'Linux'
          if (ua.includes('Android')) return 'Android'
          if (ua.includes('iPhone')) return 'iOS'
          return 'Unknown'
        }

        return (
          <Tooltip title={userAgent}>
            <Space direction="vertical" size={0}>
              <Tag color="green">{getBrowser(userAgent)}</Tag>
              <span style={{ fontSize: '12px', color: '#666' }}>
                {getOS(userAgent)}
              </span>
            </Space>
          </Tooltip>
        )
      },
    },
    {
      title: 'Referer',
      dataIndex: 'referer',
      key: 'referer',
      width: 150,
      render: (referer: string) => {
        if (!referer) return <span style={{ color: '#999' }}>Direct</span>
        try {
          const url = new URL(referer)
          return (
            <Tooltip title={referer}>
              <Tag color="orange">{url.hostname}</Tag>
            </Tooltip>
          )
        } catch {
          return (
            <Tooltip title={referer}>
              <Tag color="orange">Invalid URL</Tag>
            </Tooltip>
          )
        }
      },
    }
  ]

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Title level={2}>Access Logs</Title>
        <p>Detailed access logs for all redirect requests</p>
      </div>

      {/* Filters */}
      <Card style={{ marginBottom: '24px' }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={6}>
            <label style={{ marginRight: '8px' }}>Link:</label>
            <Select
              style={{ width: '100%' }}
              placeholder="All links"
              value={filters.link_id || undefined}
              onChange={(value) => handleFilterChange('link_id', value || '')}
              allowClear
            >
              {linksData?.data?.map((link: any) => (
                <Option key={link.link_id} value={link.link_id}>
                  {link.link_id} - {link.business_unit}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <label style={{ marginRight: '8px' }}>Country:</label>
            <Select
              style={{ width: '100%' }}
              placeholder="All countries"
              value={filters.country || undefined}
              onChange={(value) => handleFilterChange('country', value || '')}
              allowClear
            >
              <Option value="US">🇺🇸 United States</Option>
              <Option value="CN">🇨🇳 China</Option>
              <Option value="GB">🇬🇧 United Kingdom</Option>
              <Option value="DE">🇩🇪 Germany</Option>
              <Option value="FR">🇫🇷 France</Option>
              <Option value="IT">🇮🇹 Italy</Option>
              <Option value="AU">🇦🇺 Australia</Option>
              <Option value="TW">🇹🇼 Taiwan</Option>
              <Option value="LOCAL">🏠 Local</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <label style={{ marginRight: '8px' }}>IP Address:</label>
            <Input
              placeholder="Filter by IP"
              value={filters.ip}
              onChange={(e) => handleFilterChange('ip', e.target.value)}
              prefix={<SearchOutlined />}
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Space>
              <Button 
                onClick={clearFilters} 
                disabled={!filters.link_id && !filters.country && !filters.ip}
              >
                Clear Filters
              </Button>
              <Button type="primary" icon={<ReloadOutlined />} onClick={() => refetch()}>
                Refresh
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Summary Stats */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#1890ff' }}>
                {logsData?.total?.toLocaleString() || 0}
              </div>
              <div style={{ color: '#666' }}>Total Records</div>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#52c41a' }}>
                {logsData?.total_pages || 0}
              </div>
              <div style={{ color: '#666' }}>Total Pages</div>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#fa8c16' }}>
                {filters.page}
              </div>
              <div style={{ color: '#666' }}>Current Page</div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* Access Logs Table */}
      <Card>
        <Table
          columns={columns}
          dataSource={logsData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: logsData?.page || 1,
            pageSize: logsData?.page_size || 20,
            total: logsData?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `${range[0]}-${range[1]} of ${total} items`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
          size="small"
        />
      </Card>
    </div>
  )
}

export default AccessLogs