import { GithubOutlined } from '@ant-design/icons';
import { DefaultFooter } from '@ant-design/pro-components';
import React from 'react';

const Footer: React.FC = () => {
  return (
    <DefaultFooter
      style={{
        background: 'none',
      }}
      links={[
        {
          key: 'P14-AI Comment Moderation',
          title: 'P14-AI Comment Moderation',
          href: '',
          blankTarget: true,
        },
        {
          key: 'github',
          title: <GithubOutlined />,
          href: 'https://github.com/unsw-cse-comp99-3900/capstone-project-2025-t1-25t1-9900-t18b-brioche',
          blankTarget: true,
        },
        {
          key: '9900_T18B_BRIOCHE',
          title: '9900_T18B_BRIOCHE',
          href: '',
          blankTarget: true,
        },
      ]}
    />
  );
};

export default Footer;
