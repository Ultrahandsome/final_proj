import React from 'react';
import { Dropdown, Menu, Button, message } from 'antd';
import { ExportOutlined, DownOutlined } from '@ant-design/icons';

interface ExportFileProps {
  selectedIds?: string[];
}

const ExportFile: React.FC<ExportFileProps> = ({ selectedIds = [] }) => {
  const handleExport = async (format: 'csv' | 'excel' | 'tsv') => {
    const token = localStorage.getItem('token') || '';
    const endpoint = `/api/export/${format}`;
    let filename = format === 'csv' ? 'comments.csv' : 'comments.xlsx';
    if (format === 'tsv') {
      filename = 'comments.tsv';
    }

    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Token': token,
        },
        body: JSON.stringify({ ids: selectedIds }),
      });

      if (!response.ok) {
        throw new Error('Export failed');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export error:', error);
      message.error('Export failed, please try again later.');
    }
  };

  const menu = (
    <Menu
      onClick={({ key }) => handleExport(key as 'csv' | 'excel')}
      items={[
        { key: 'csv', label: 'Export as CSV' },
        { key: 'excel', label: 'Export as Excel' },
        { key: 'tsv', label: 'Export as TSV' },
      ]}
    />
  );

  return (
    <Dropdown overlay={menu}>
      <Button icon={<ExportOutlined />}>
        Export <DownOutlined />
      </Button>
    </Dropdown>
  );
};

export default ExportFile;
