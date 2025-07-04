import React, { useState } from 'react'
import { Modal, Form, Input, Select, InputNumber, Button, Space, Divider, Alert, Card } from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { batchApi } from '@/services/api'
import { message } from 'antd'

interface QuickCreateModalProps {
  visible: boolean
  onClose: () => void
}

const countryOptions = [
  { label: '🌍 All Countries', value: 'ALL' },
  { label: '🇺🇸 United States', value: 'US' },
  { label: '🇨🇦 Canada', value: 'CA' },
  { label: '🇬🇧 United Kingdom', value: 'UK' },
  { label: '🇩🇪 Germany', value: 'DE' },
  { label: '🇫🇷 France', value: 'FR' },
  { label: '🇯🇵 Japan', value: 'JP' },
  { label: '🇨🇳 China', value: 'CN' },
  { label: '🇮🇳 India', value: 'IN' },
  { label: '🇧🇷 Brazil', value: 'BR' },
  { label: '🇦🇺 Australia', value: 'AU' },
]

const QuickCreateModal: React.FC<QuickCreateModalProps> = ({ visible, onClose }) => {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)
      
      // Transform form values to API format
      const linkData = {
        business_unit: values.business_unit,
        network: values.network,
        total_cap: values.total_cap || 0,
        backup_url: values.backup_url || '',
        targets: values.targets?.map((target: any) => ({
          url: target.url,
          weight: target.weight || 100,
          cap: target.cap || 0,
          countries: target.countries || [],
          param_mapping: {},
          static_params: {},
        })) || []
      }

      const response = await batchApi.createLinks({
        links: [linkData]
      })
      
      message.success('Link and targets created successfully!')
      form.resetFields()
      onClose()
      
      // Navigate to the created link
      if (response.data?.created?.[0]?.link_id) {
        navigate(`/links/${response.data.created[0].link_id}`)
      }
    } catch (error: any) {
      message.error(error.response?.data?.error || 'Failed to create link')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="Quick Create Link with Targets"
      open={visible}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={loading}
      width={800}
      okText="Create Link & Targets"
    >
      <Alert
        message="Create a link and its target URLs in one step"
        description="You can add multiple target URLs with different weights and geographic restrictions"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          network: 'mi',
          total_cap: 0,
          targets: [{ weight: 100, cap: 0 }]
        }}
      >
        {/* Link Information */}
        <Divider orientation="left">Link Information</Divider>
        
        <Form.Item
          name="business_unit"
          label="Business Unit"
          rules={[{ required: true, message: 'Please select business unit' }]}
        >
          <Select>
            <Select.Option value="bu01">bu01 - Africa Business</Select.Option>
            <Select.Option value="bu02">bu02 - Latin America Business</Select.Option>
            <Select.Option value="bu03">bu03 - Asia Business</Select.Option>
            <Select.Option value="bu04">bu04 - Europe Business</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          name="network"
          label="Network (Channel)"
          rules={[{ required: true, message: 'Please enter network channel' }]}
        >
          <Input placeholder="e.g. mi, google, fb, tiktok" />
        </Form.Item>

        <Form.Item
          name="total_cap"
          label="Total Link Cap (0 for unlimited)"
          tooltip="Maximum number of hits allowed for this link"
        >
          <InputNumber
            min={0}
            style={{ width: '100%' }}
            placeholder="Enter 0 for unlimited"
          />
        </Form.Item>

        <Form.Item
          name="backup_url"
          label="Backup URL"
          tooltip="URL to redirect when cap is reached or no targets available"
          rules={[
            { type: 'url', message: 'Please enter a valid URL' }
          ]}
        >
          <Input placeholder="https://example.com/backup" />
        </Form.Item>

        {/* Target URLs */}
        <Divider orientation="left">Target URLs</Divider>
        
        <Form.List name="targets">
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name, ...restField }, index) => (
                <Card
                  key={key}
                  size="small"
                  title={`Target ${index + 1}`}
                  style={{ marginBottom: 16 }}
                  extra={
                    fields.length > 1 && (
                      <Button
                        type="text"
                        danger
                        icon={<MinusCircleOutlined />}
                        onClick={() => remove(name)}
                      >
                        Remove
                      </Button>
                    )
                  }
                >
                  <Form.Item
                    {...restField}
                    name={[name, 'url']}
                    label="Target URL"
                    rules={[
                      { required: true, message: 'Please enter target URL' },
                      { type: 'url', message: 'Please enter a valid URL' },
                    ]}
                  >
                    <Input placeholder="https://example.com/landing" />
                  </Form.Item>

                  <Space style={{ width: '100%' }} size="large">
                    <Form.Item
                      {...restField}
                      name={[name, 'weight']}
                      label="Weight (%)"
                      rules={[{ required: true, message: 'Please enter weight' }]}
                      style={{ marginBottom: 0 }}
                    >
                      <InputNumber
                        min={1}
                        max={1000}
                        style={{ width: 120 }}
                      />
                    </Form.Item>

                    <Form.Item
                      {...restField}
                      name={[name, 'cap']}
                      label="Cap (0 = unlimited)"
                      style={{ marginBottom: 0 }}
                    >
                      <InputNumber
                        min={0}
                        style={{ width: 150 }}
                      />
                    </Form.Item>

                    <Form.Item
                      {...restField}
                      name={[name, 'countries']}
                      label="Allowed Countries"
                      style={{ marginBottom: 0, flex: 1 }}
                    >
                      <Select
                        mode="multiple"
                        placeholder="All countries if empty"
                        options={countryOptions}
                        allowClear
                      />
                    </Form.Item>
                  </Space>
                </Card>
              ))}
              
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add({ weight: 100, cap: 0 })}
                  block
                  icon={<PlusOutlined />}
                >
                  Add Target URL
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Form>
    </Modal>
  )
}

export default QuickCreateModal