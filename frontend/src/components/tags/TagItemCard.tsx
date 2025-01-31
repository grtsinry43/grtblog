'use client';

import React from 'react';
import { TagIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import { clsx } from 'clsx';
import { motion } from 'framer-motion';

const TagItemCard = ({ name, count, index }: { name: string, count: number, index: number }) => {
  const colors = ['bg-blue-100', 'bg-green-100', 'bg-yellow-100', 'bg-red-100', 'bg-indigo-100', 'bg-purple-100', 'bg-pink-100'];
  const color = colors[index % colors.length];
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ type: 'spring', stiffness: 100, mass: 0.8, delay: index * 0.05 }}
      whileHover={{
        scale: 1.05,
        transition: { duration: 0.2 },
      }}
    >
      <Link href={`/tags/${name}`} className={clsx('flex h-12 bg-opacity-20 backdrop-blur', color)} style={{
        borderRadius: '0.5rem',
        backdropFilter: 'blur(10px)',
        margin: '0.5rem',
        padding: '0 1rem',
        alignItems: 'center',
      }}>
        <TagIcon className="h-6 w-6 text-gray-400" aria-hidden="true" />
        <span className="ml-2">{name}</span>
        <span className="ml-4 text-sm text-gray-400 dark:text-gray-500">{count}</span>
      </Link>
    </motion.div>
  );
};

export default TagItemCard;
