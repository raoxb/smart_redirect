import React from 'react'
import { Modal, Form, Input, Select, InputNumber, Alert } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useCreateLink } from '@/hooks/useApi'
import type { CreateLinkRequest } from '@/types/api'

interface CreateLinkModalProps {
  visible: boolean
  onClose: () => void
}

const CreateLinkModal: React.FC<CreateLinkModalProps> = ({ visible, onClose }) => {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const createMutation = useCreateLink()

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data: CreateLinkRequest = {
        business_unit: values.business_unit,
        network: values.network,
        total_cap: values.total_cap || 0,
        backup_url: values.backup_url || '',
      }
      
      const response = await createMutation.mutateAsync(data)
      form.resetFields()
      onClose()
      // Navigate to the link detail page to add targets
      if (response.data?.link_id) {
        navigate(`/links/${response.data.link_id}`)
      }
    } catch (error) {
      // Error handling is done in the hook
    }
  }

  return (
    <Modal
      title="Create New Link"
      open={visible}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={createMutation.isPending}
      width={600}
    >
      <Alert
        message="After creating the link, you'll be redirected to add target URLs"
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
        }}
      >
        <Form.Item
          name="business_unit"
          label="Business Unit"
          rules={[{ required: true, message: 'Please select business unit' }]}
        >
          <Select>
            <Select.Option value="bu01">bu01 - 非洲业务</Select.Option>
            <Select.Option value="bu02">bu02 - 拉美业务</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          name="network"
          label="Network (Channel)"
          rules={[{ required: true, message: 'Please enter network channel' }]}
        >
          <Input placeholder="e.g. mi, google, fb" />
        </Form.Item>

        <Form.Item
          name="total_cap"
          label="Total Cap (0 for unlimited)"
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
      </Form>
    </Modal>
  )
}

export default CreateLinkModal