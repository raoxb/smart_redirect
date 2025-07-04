import React, { useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Popconfirm,
  Tag,
  Tooltip,
  Row,
  Col,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
  FileTextOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/services/api'
import type { ColumnsType } from 'antd/es/table'

const { TextArea } = Input
const { Option } = Select

interface Template {
  id: number
  name: string
  description: string
  business_unit: string
  network: string
  total_cap: number
  backup_url: string
  targets: TemplateTarget[]
  created_at: string
  updated_at: string
}

interface TemplateTarget {
  url: string
  weight: number
  cap: number
  countries: string[]
  param_mapping: Record<string, string>
  static_params: Record<string, string>
}

interface CreateTemplateRequest {
  name: string
  description: string
  business_unit: string
  network: string
  total_cap: number
  backup_url: string
  targets: TemplateTarget[]
}

const Templates: React.FC = () => {
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()

  // Fetch templates
  const { data: templatesData, isLoading } = useQuery({
    queryKey: ['templates'],
    queryFn: () => api.get('/templates').then(res => res.data),
  })

  // Create template mutation
  const createTemplateMutation = useMutation({
    mutationFn: (data: CreateTemplateRequest) => api.post('/templates', data),
    onSuccess: () => {
      message.success('Template created successfully')
      setIsCreateModalVisible(false)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['templates'] })
    },
    onError: () => {
      message.error('Failed to create template')
    },
  })

  // Update template mutation
  const updateTemplateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateTemplateRequest }) =>
      api.put(`/templates/${id}`, data),
    onSuccess: () => {
      message.success('Template updated successfully')
      setEditingTemplate(null)
      form.resetFields()
      queryClient.invalidateQueries({ queryKey: ['templates'] })
    },
    onError: () => {
      message.error('Failed to update template')
    },
  })

  // Delete template mutation
  const deleteTemplateMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/templates/${id}`),
    onSuccess: () => {
      message.success('Template deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['templates'] })
    },
    onError: () => {
      message.error('Failed to delete template')
    },
  })

  // Create links from template mutation
  const createLinksFromTemplateMutation = useMutation({
    mutationFn: ({ templateId, count }: { templateId: number; count: number }) =>
      api.post('/templates/create-links', { template_id: templateId, count }),
    onSuccess: (response) => {
      const createdCount = response.data?.created_count || 0
      message.success(`Created ${createdCount} links from template`)
      queryClient.invalidateQueries({ queryKey: ['links'] })
    },
    onError: () => {
      message.error('Failed to create links from template')
    },
  })

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      
      if (editingTemplate) {
        updateTemplateMutation.mutate({ id: editingTemplate.id, data: values })
      } else {
        createTemplateMutation.mutate(values)
      }
    } catch (error) {
      console.error('Form validation failed:', error)
    }
  }

  const handleEdit = (template: Template) => {
    setEditingTemplate(template)
    form.setFieldsValue({
      ...template,
      targets: template.targets || [],
    })
    setIsCreateModalVisible(true)
  }

  const handleCreateLinks = (templateId: number) => {
    Modal.confirm({
      title: 'Create Links from Template',
      content: (
        <div>
          <p>How many links would you like to create from this template?</p>
          <InputNumber
            min={1}
            max={100}
            defaultValue={1}
            onChange={(value) => {
              // Store the count value for use in the confirm handler
              ;(Modal as any)._linkCount = value
            }}
          />
        </div>
      ),
      onOk: () => {
        const count = (Modal as any)._linkCount || 1
        createLinksFromTemplateMutation.mutate({ templateId, count })
      },
    })
  }

  const columns: ColumnsType<Template> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Template) => (
        <div>
          <div style={{ fontWeight: 600 }}>{text}</div>
          <div style={{ fontSize: '12px', color: '#666' }}>{record.description}</div>
        </div>
      ),
    },
    {
      title: 'Business Unit',
      dataIndex: 'business_unit',
      key: 'business_unit',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: 'Network',
      dataIndex: 'network',
      key: 'network',
      render: (text: string) => <Tag color="green">{text}</Tag>,
    },
    {
      title: 'Targets',
      dataIndex: 'targets',
      key: 'targets',
      render: (targets: TemplateTarget[]) => (
        <span>{targets?.length || 0} targets</span>
      ),
    },
    {
      title: 'Total Cap',
      dataIndex: 'total_cap',
      key: 'total_cap',
      render: (cap: number) => cap || 'Unlimited',
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString(),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record: Template) => (
        <Space>
          <Tooltip title="Create Links">
            <Button
              type="primary"
              icon={<LinkOutlined />}
              size="small"
              onClick={() => handleCreateLinks(record.id)}
            />
          </Tooltip>
          <Tooltip title="Edit Template">
            <Button
              icon={<EditOutlined />}
              size="small"
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="Duplicate Template">
            <Button
              icon={<CopyOutlined />}
              size="small"
              onClick={() => {
                const duplicate = { ...record }
                delete (duplicate as any).id
                duplicate.name = `${record.name} (Copy)`
                createTemplateMutation.mutate(duplicate)
              }}
            />
          </Tooltip>
          <Popconfirm
            title="Are you sure you want to delete this template?"
            onConfirm={() => deleteTemplateMutation.mutate(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              danger
              icon={<DeleteOutlined />}
              size="small"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <h1>Templates</h1>
            <p>Manage link templates for batch creation</p>
          </Col>
          <Col>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditingTemplate(null)
                form.resetFields()
                setIsCreateModalVisible(true)
              }}
            >
              Create Template
            </Button>
          </Col>
        </Row>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={templatesData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} templates`,
          }}
        />
      </Card>

      {/* Create/Edit Template Modal */}
      <Modal
        title={editingTemplate ? 'Edit Template' : 'Create Template'}
        open={isCreateModalVisible}
        onCancel={() => {
          setIsCreateModalVisible(false)
          setEditingTemplate(null)
          form.resetFields()
        }}
        onOk={handleSubmit}
        confirmLoading={createTemplateMutation.isPending || updateTemplateMutation.isPending}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            total_cap: 0,
            targets: [
              {
                url: '',
                weight: 50,
                cap: 0,
                countries: [],
                param_mapping: {},
                static_params: {},
              },
            ],
          }}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="Template Name"
                rules={[{ required: true, message: 'Please enter template name' }]}
              >
                <Input placeholder="Enter template name" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="business_unit"
                label="Business Unit"
                rules={[{ required: true, message: 'Please enter business unit' }]}
              >
                <Input placeholder="e.g., bu01" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="Description"
          >
            <TextArea
              placeholder="Enter template description"
              rows={2}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="network"
                label="Network"
                rules={[{ required: true, message: 'Please enter network' }]}
              >
                <Input placeholder="e.g., social_media" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="total_cap"
                label="Total Cap"
              >
                <InputNumber
                  style={{ width: '100%' }}
                  placeholder="0 for unlimited"
                  min={0}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="backup_url"
            label="Backup URL"
          >
            <Input placeholder="https://example.com/backup" />
          </Form.Item>

          <Form.Item label="Targets">
            <Form.List name="targets">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...restField }) => (
                    <Card
                      key={key}
                      size="small"
                      style={{ marginBottom: 16 }}
                      title={`Target ${name + 1}`}
                      extra={
                        fields.length > 1 && (
                          <Button
                            type="text"
                            danger
                            size="small"
                            onClick={() => remove(name)}
                          >
                            Remove
                          </Button>
                        )
                      }
                    >
                      <Row gutter={16}>
                        <Col span={16}>
                          <Form.Item
                            {...restField}
                            name={[name, 'url']}
                            label="URL"
                            rules={[{ required: true, message: 'URL is required' }]}
                          >
                            <Input placeholder="https://example.com/landing" />
                          </Form.Item>
                        </Col>
                        <Col span={8}>
                          <Form.Item
                            {...restField}
                            name={[name, 'weight']}
                            label="Weight"
                            rules={[{ required: true, message: 'Weight is required' }]}
                          >
                            <InputNumber
                              style={{ width: '100%' }}
                              min={1}
                              placeholder="50"
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                      
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item
                            {...restField}
                            name={[name, 'cap']}
                            label="Cap"
                          >
                            <InputNumber
                              style={{ width: '100%' }}
                              min={0}
                              placeholder="0 for unlimited"
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            {...restField}
                            name={[name, 'countries']}
                            label="Countries"
                          >
                            <Select
                              mode="tags"
                              style={{ width: '100%' }}
                              placeholder="US, UK, DE..."
                              tokenSeparators={[',']}
                            >
                              <Option value="US">United States</Option>
                              <Option value="UK">United Kingdom</Option>
                              <Option value="DE">Germany</Option>
                              <Option value="FR">France</Option>
                              <Option value="CA">Canada</Option>
                              <Option value="AU">Australia</Option>
                            </Select>
                          </Form.Item>
                        </Col>
                      </Row>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    Add Target
                  </Button>
                </>
              )}
            </Form.List>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Templates