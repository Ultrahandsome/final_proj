import React, { useState } from 'react';
import { InboxOutlined } from '@ant-design/icons';
import { Upload, message, Card, Typography, Progress, Space } from 'antd';
import type { UploadProps } from 'antd';
import { request } from 'umi';
import { PageContainer } from '@ant-design/pro-components';

const { Dragger } = Upload;
const { Text } = Typography;

const DataImport: React.FC = () => {
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [fileList, setFileList] = useState<any[]>([]);

  const handleUpload = async (options: any) => {
    const { file, onSuccess, onError, onProgress } = options;

    // Determine file extension
    const fileName = file.name;
    const fileExtension = fileName.substring(fileName.lastIndexOf('.') + 1).toLowerCase();

    // Determine API endpoint based on file type
    let uploadUrl;
    if (fileExtension === 'csv') {
      uploadUrl = '/api/upload/csv';
    } else if (fileExtension === 'xlsx') {
      uploadUrl = '/api/upload/excel';
    } else if (fileExtension === 'tsv') {
      uploadUrl = '/api/upload/tsv';
    } else {
      message.error(`Unsupported file type: ${fileExtension}`);
      onError(new Error(`Unsupported file type: ${fileExtension}`));
      return;
    }

    const formData = new FormData();
    formData.append('file', file);

    try {
      setUploading(true);

      const res = await request(uploadUrl, {
        method: 'POST',
        data: formData,
        requestType: 'form',
        // Track upload progress
        onUploadProgress: (progressEvent) => {
          const { loaded, total } = progressEvent;
          const percent = Math.round((loaded * 100) / (total ?? 100));
          setUploadProgress(percent);
          onProgress({ percent });
        },
      });

      setUploading(false);
      setUploadProgress(0);
      message.success(`${file.name} uploaded successfully.`);
      onSuccess(res, file);
    } catch (error) {
      setUploading(false);
      setUploadProgress(0);
      message.error(`${file.name} upload failed.`);
      onError(error);
    }
  };

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    fileList,
    customRequest: handleUpload,
    onChange(info) {
      const { status } = info.file;

      // Update fileList
      setFileList(info.fileList.slice(-1)); // Only keep the latest file

      if (status === 'done') {
        message.success(`${info.file.name} file uploaded successfully.`);
      } else if (status === 'error') {
        message.error(`${info.file.name} file upload failed.`);
      }
    },
    beforeUpload(file) {
      const fileName = file.name;
      const fileExtension = fileName.substring(fileName.lastIndexOf('.') + 1).toLowerCase();

      // Check if file type is supported
      const isCSVOrXLSX = ['csv', 'xlsx', 'tsv'].includes(fileExtension);
      if (!isCSVOrXLSX) {
        message.error('Only CSV, TSV, and XLSX files allowed!');
        return Upload.LIST_IGNORE;
      }

      // Check file size (optional, adjust as needed)
      const isLessThan50MB = file.size / 1024 / 1024 < 50;
      if (!isLessThan50MB) {
        message.error('File must be smaller than 50MB!');
        return Upload.LIST_IGNORE;
      }

      return true;
    },
    onRemove() {
      setFileList([]);
      return true;
    },
  };

  return (
    <PageContainer>
      <Card
        style={{
          borderRadius: 8,
        }}
      >
        <div
          style={{
            fontSize: '20px',
            marginBottom: 16,
          }}
        >
          Comments Import
        </div>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Text>Upload a CSV or Excel (XLSX) file to begin.</Text>

          <Dragger {...uploadProps}>
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">Click or drag file to this area to upload</p>
            <p className="ant-upload-hint">
              Supports CSV, TSV, and Excel (XLSX) files only. Maximum file size: 50MB.
            </p>
          </Dragger>

          {uploading && (
            <Progress percent={uploadProgress} status="active" style={{ marginTop: 16 }} />
          )}
        </Space>
      </Card>
    </PageContainer>
  );
};

export default DataImport;
