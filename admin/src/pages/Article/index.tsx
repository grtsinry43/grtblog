import React from 'react';
import { PageContainer, ProTable } from '@ant-design/pro-components';

const Index = () => {
  const columns = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '封面',
      dataIndex: 'cover',
      key: 'cover',
      valueType: 'image',
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
    },
    {
      title: '操作',
      valueType: 'option',
      render: () => [
        <a key="edit"> 编辑 </a>,
        <a key="delete"> 删除 </a>,
      ],
    },
  ];

  return (
    <PageContainer title={'文章列表'}>
      <ProTable
        columns={columns}
        request={async () => {
          return {
            data: [
              {
                title: '标题1',
                cover: 'https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg',
                category: '分类1',
                tags: '标签1',
                status: '状态1',
                createdAt: '2024-10-09 12:00:00',
                updatedAt: '2024-10-09 12:00:00',
              },
              {
                title: '标题2',
                cover: 'https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg',
                category: '分类2',
                tags: '标签2',
                status: '状态2',
                createdAt: '2024-10-09 12:00:00',
                updatedAt: '2024-10-09 12:00:00',
              },
            ],
            success: true,
          };
        }}
      />
    </PageContainer>
  );
};

export default Index;
