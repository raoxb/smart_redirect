import React, { useState } from 'react'
import { Modal, Upload, Button, Alert, Typography, Table, Tag } from 'antd'
import { InboxOutlined, DownloadOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { useBatchImport } from '@/hooks/useApi'
import { batchApi } from '@/services/api'
import { downloadFile } from '@/utils/format'

const { Dragger } = Upload
const { Text, Link } = Typography

interface ImportModalProps {
  visible: boolean
  onClose: () => void
}

const ImportModal: React.FC<ImportModalProps> = ({ visible, onClose }) => {
  const [file, setFile] = useState<File | null>(null)
  const [importResults, setImportResults] = useState<any>(null)
  const importMutation = useBatchImport()

  const handleImport = async () => {
    if (!file) return

    const response = await importMutation.mutateAsync(file)
    setImportResults(response.data)
  }

  const handleDownloadTemplate = async () => {
    try {
      const response = await batchApi.exportCSV()
      downloadFile(response.data, 'smart_redirect_template.csv')
    } catch (error) {
      // Handle error
    }
  }

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    accept: '.csv',
    beforeUpload: (file) => {
      setFile(file)
      return false
    },
    onRemove: () => {
      setFile(null)
    },
  }

  const handleClose = () => {
    setFile(null)
    setImportResults(null)
    onClose()
  }

  return (
    <Modal
      title="Import Links from CSV"
      open={visible}
      onOk={handleImport}
      onCancel={handleClose}
      confirmLoading={importMutation.isPending}
      width={700}
      okText="Import"
      okButtonProps={{ disabled: !file || !!importResults }}
    >
      {!importResults ? (
        <>
          <Alert
            message="CSV Format"
            description={
              <div>
                <Text>Your CSV file should include the following columns:</Text>
                <ul style={{ marginTop: 8, marginBottom: 8 }}>
                  <li><code>business_unit</code> - Business unit code (e.g., bu01, bu02)</li>
                  <li><code>network</code> - Network channel (e.g., mi, google)</li>
                  <li><code>total_cap</code> - Total cap limit (0 for unlimited)</li>
                  <li><code>backup_url</code> - Backup URL (optional)</li>
                  <li><code>target_url</code> - Target redirect URL</li>
                  <li><code>weight</code> - Target weight for traffic distribution</li>
                  <li><code>cap</code> - Target cap limit</li>
                  <li><code>countries</code> - Semicolon-separated country codes (e.g., US;CA;UK)</li>
                </ul>
                <Link onClick={handleDownloadTemplate}>
                  <DownloadOutlined /> Download template CSV
                </Link>
              </div>
            }
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <Dragger {...uploadProps}>
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">Click or drag CSV file to this area to upload</p>
            <p className="ant-upload-hint">
              Support for a single CSV file upload. The file should follow the format described above.
            </p>
          </Dragger>
        </>
      ) : (
        <div>
          <Alert
            message="Import Complete"
            description={`Successfully imported ${importResults.success.length} links. ${importResults.errors.length} errors occurred.`}
            type={importResults.errors.length > 0 ? 'warning' : 'success'}
            showIcon
            style={{ marginBottom: 16 }}
          />

          {importResults.success.length > 0 && (
            <div style={{ marginBottom: 16 }}>
              <Text strong>Successfully Imported:</Text>
              <Table
                size="small"
                dataSource={importResults.success}
                columns={[
                  { title: 'Row', dataIndex: 'index', key: 'index' },
                  { title: 'Link ID', dataIndex: 'link_id', key: 'link_id' },
                  { title: 'URL', dataIndex: 'link_url', key: 'link_url' },
                ]}
                pagination={false}
                style={{ marginTop: 8 }}
              />
            </div>
          )}

          {importResults.errors.length > 0 && (
            <div>
              <Text strong>Errors:</Text>
              <Table
                size="small"
                dataSource={importResults.errors}
                columns={[
                  { title: 'Row', dataIndex: 'index', key: 'index' },
                  { 
                    title: 'Error', 
                    dataIndex: 'message', 
                    key: 'message',
                    render: (msg: string) => <Text type="danger">{msg}</Text>
                  },
                ]}
                pagination={false}
                style={{ marginTop: 8 }}
              />
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}

export default ImportModal