'use client';

import React, { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';
import { motion, AnimatePresence } from 'framer-motion';
import Link from 'next/link';
import { Avatar, Button } from '@radix-ui/themes';
import { GitHubLogoIcon, HamburgerMenuIcon, MagnifyingGlassIcon, MoonIcon, SunIcon } from '@radix-ui/react-icons';
import styles from '@/styles/NavBarMobile.module.scss';

const NavBarMobile = () => {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [showPanel, setShowPanel] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const navItems = [
    {
      name: '首页',
      href: '/',
      children: [
        { name: '留言板', href: '/comments' },
        { name: '友链', href: '/friends' },
      ],
    },
    { name: '关于', href: '/about' },
    { name: '分类', href: '/categories' },
    { name: '文章', href: '/posts' },
    { name: '项目', href: '/projects' },
    { name: '标签', href: '/tags' },
  ];

  const togglePanel = () => {
    setShowPanel(!showPanel);
  };

  return (
    <div className="fixed w-full z-10">
      <motion.div
        initial={{ y: -100 }}
        animate={{ y: 0 }}
        transition={{ type: 'spring', stiffness: 120, damping: 20 }}
      >
        <nav
          className="w-full bg-background/80 backdrop-blur-xl text-foreground p-4 sticky top-0 z-10 shadow-md transition-all duration-200">
          <div className="max-w-7xl mx-auto flex justify-between items-center">
            <button onClick={togglePanel}
                    className="px-4 py-2 bg-primary text-primary-foreground rounded-md transition-colors duration-200 hover:bg-primary/90">
              <HamburgerMenuIcon className="w-5 h-5" />
            </button>
            <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }}
                        transition={{ type: 'spring', stiffness: 260, damping: 20 }}>
              <Avatar
                size="3"
                radius="large"
                src="https://dogeoss.grtsinry43.com/img/author.jpeg"
                fallback="A"
                className={styles.avatar}
              />
            </motion.div>
            <div className={styles.navbarContent}>
              <div className={styles.searchWrapper}>
                <motion.div whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.95 }}>
                  <div className={styles.search}>
                    <MagnifyingGlassIcon className={styles.searchIcon} />
                  </div>
                </motion.div>
              </div>
              <div className={styles.githubWrapper}>
                <motion.div whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.95 }}>
                  <a
                    href="https://github.com/grtsinry43"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.githubLink}
                  >
                    <GitHubLogoIcon />
                  </a>
                </motion.div>
              </div>
              <div className={styles.themeToggleWrapper}>
                <motion.div whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.95 }}>
                  {mounted && (
                    <div
                      onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
                      className={styles.themeToggle}
                    >
                      {resolvedTheme === 'dark' ? <SunIcon /> : <MoonIcon />}
                    </div>
                  )}
                </motion.div>
              </div>
              <div className={styles.loginButtonWrapper}>
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button variant="soft" className={styles.loginButton}>
                    登录
                  </Button>
                </motion.div>
              </div>
            </div>
          </div>
        </nav>
      </motion.div>
      <AnimatePresence>
        {showPanel && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.2 }}
          >
            <div
              className="fixed inset-x-0 bg-background/95 backdrop-blur-md shadow-lg rounded-b-lg overflow-hidden z-20">
              <div className="max-h-[70vh] overflow-y-auto p-4 space-y-4">
                {navItems.map((item) => (
                  <div key={item.name} className="space-y-2">
                    <Link href={item.href}
                          className="block text-lg font-medium text-foreground hover:text-primary transition-colors duration-200">
                      {item.name}
                    </Link>
                    {item.children && (
                      <div className="pl-4 space-y-2">
                        {item.children.map((child) => (
                          <Link
                            key={child.name}
                            href={child.href}
                            className="block text-muted-foreground hover:text-primary transition-colors duration-200 pl-2"
                            onClick={() => setShowPanel(false)}
                          >
                            {child.name}
                          </Link>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      <motion.div initial={{ scale: 0.8, opacity: 0 }} animate={{ scale: 1, opacity: 1 }}
                  transition={{ type: 'spring', stiffness: 100, damping: 10, delay: 0.2 }}>
        <div
          className="fixed inset-0 bg-gradient-radial from-primary/10 to-background/10 -z-10 pointer-events-none"></div>
      </motion.div>
    </div>
  );
};

export default NavBarMobile;