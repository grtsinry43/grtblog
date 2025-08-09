'use client';

import React from 'react';
import styles from '@/styles/Home.module.scss';
import {Avatar, Container, HoverCard, Text, Flex, Box, Heading} from '@radix-ui/themes';
import {CodeIcon, GitHubLogoIcon} from '@radix-ui/react-icons';
import {clsx} from 'clsx';
import {motion} from 'framer-motion';
import {useWebsiteInfo} from "@/app/website-info-provider";
import GridPattern from "@/app/home/GridPattern";

const AuthorCard = () => {
    const websiteInfo = useWebsiteInfo();
    return (
        <motion.div
            initial={{opacity: 0}}
            animate={{opacity: 1}}
            transition={{duration: 0.5}}
        >
            <div style={{
                transition: 'all 0.5s',
            }}
                 className={styles.bannerCard}>
                <div className="relative w-full max-w-md">
                    {/* 增强的网格背景 */}
                    <GridPattern
                        width={30}
                        height={30}
                        x={-1}
                        y={-1}
                        strokeDasharray={'4 4'}
                        className="[mask-image:radial-gradient(ellipse_at_center,white,transparent_70%)] opacity-30" />

                    {/* 光晕背景层 - 固定在视口中，不影响布局 */}
                    <div className="fixed inset-0 pointer-events-none -z-10">
                        {/* 主光晕 - 卡片左上角 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                left: '35%',
                                top: '8%',
                            }}
                            animate={{
                                scale: [1, 1.1, 1],
                                opacity: [0.12, 0.22, 0.12],
                            }}
                            transition={{
                                duration: 4,
                                repeat: Infinity,
                                ease: "easeInOut"
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-3xl"
                                style={{
                                    width: '200px',
                                    height: '200px',
                                    left: '-100px',
                                    top: '-100px',
                                    background: 'radial-gradient(circle, #3b82f6 0%, #1d4ed8 40%, transparent 70%)'
                                }}
                            />
                        </motion.div>
                        
                        {/* 辅助光晕 - 卡片右侧 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                right: '30%',
                                top: '20%',
                            }}
                            animate={{
                                scale: [1, 1.15, 1],
                                opacity: [0.1, 0.18, 0.1],
                            }}
                            transition={{
                                duration: 5,
                                repeat: Infinity,
                                ease: "easeInOut",
                                delay: 1
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-3xl"
                                style={{
                                    width: '150px',
                                    height: '150px',
                                    left: '-75px',
                                    top: '-75px',
                                    background: 'radial-gradient(circle, #f97316 0%, #ea580c 40%, transparent 70%)'
                                }}
                            />
                        </motion.div>
                        
                        {/* 底部光晕 - 卡片下方 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                left: '42%',
                                top: '40%',
                            }}
                            animate={{
                                scale: [1, 1.2, 1],
                                opacity: [0.08, 0.16, 0.08],
                            }}
                            transition={{
                                duration: 6,
                                repeat: Infinity,
                                ease: "easeInOut",
                                delay: 2
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-3xl"
                                style={{
                                    width: '180px',
                                    height: '180px',
                                    left: '-90px',
                                    top: '-90px',
                                    background: 'radial-gradient(circle, #a855f7 0%, #7c3aed 40%, transparent 70%)'
                                }}
                            />
                        </motion.div>

                        {/* 装饰小光点 - 卡片周围 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                left: '32%',
                                top: '15%',
                            }}
                            animate={{
                                x: [0, 8, 0],
                                y: [0, -5, 0],
                                opacity: [0.15, 0.3, 0.15],
                            }}
                            transition={{
                                duration: 3,
                                repeat: Infinity,
                                ease: "easeInOut"
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-2xl"
                                style={{
                                    width: '50px',
                                    height: '50px',
                                    left: '-25px',
                                    top: '-25px',
                                    background: 'radial-gradient(circle, #06b6d4 0%, transparent 60%)'
                                }}
                            />
                        </motion.div>
                        
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                right: '35%',
                                top: '35%',
                            }}
                            animate={{
                                x: [0, -6, 0],
                                y: [0, 6, 0],
                                opacity: [0.2, 0.35, 0.2],
                            }}
                            transition={{
                                duration: 4,
                                repeat: Infinity,
                                ease: "easeInOut",
                                delay: 1.5
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-xl"
                                style={{
                                    width: '40px',
                                    height: '40px',
                                    left: '-20px',
                                    top: '-20px',
                                    background: 'radial-gradient(circle, #10b981 0%, transparent 60%)'
                                }}
                            />
                        </motion.div>

                        {/* 星星般的微光效果 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                left: '48%',
                                top: '12%',
                            }}
                            animate={{
                                scale: [1, 1.4, 1],
                                opacity: [0.2, 0.45, 0.2],
                            }}
                            transition={{
                                duration: 2.5,
                                repeat: Infinity,
                                ease: "easeInOut"
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-lg"
                                style={{
                                    width: '20px',
                                    height: '20px',
                                    left: '-10px',
                                    top: '-10px',
                                    background: 'radial-gradient(circle, #fbbf24 0%, transparent 70%)'
                                }}
                            />
                        </motion.div>
                        
                        {/* 额外的装饰光点 */}
                        <motion.div
                            className="absolute w-0 h-0 pointer-events-none"
                            style={{
                                right: '25%',
                                top: '10%',
                            }}
                            animate={{
                                scale: [1, 1.3, 1],
                                opacity: [0.1, 0.25, 0.1],
                            }}
                            transition={{
                                duration: 3.5,
                                repeat: Infinity,
                                ease: "easeInOut",
                                delay: 0.8
                            }}
                        >
                            <div 
                                className="absolute rounded-full blur-xl"
                                style={{
                                    width: '30px',
                                    height: '30px',
                                    left: '-15px',
                                    top: '-15px',
                                    background: 'radial-gradient(circle, #ec4899 0%, transparent 60%)'
                                }}
                            />
                        </motion.div>
                    </div>

                    <motion.div
                        initial={{scale: 0}}
                        animate={{scale: 1}}
                        transition={{type: 'spring', stiffness: 260, damping: 20}}
                        drag
                        dragConstraints={{
                            top: -1,
                            left: -1,
                            right: 1,
                            bottom: 1,
                        }}
                        className="relative z-10"
                    >
                        <Container
                            className={clsx(
                                "backdrop-blur-2xl backdrop-saturate-150",
                                "bg-gradient-to-br from-white/60 via-white/40 to-white/30",
                                "dark:from-zinc-900/20 dark:via-zinc-900/15 dark:to-zinc-900/10",
                                "p-6 md:p-8 rounded-xl",
                                "border border-white/30 dark:border-white/15",
                                "shadow-2xl shadow-black/20 dark:shadow-black/40",
                                "hover:shadow-3xl hover:border-white/40 dark:hover:border-white/20",
                                "transition-all duration-500 ease-out",
                                "before:absolute before:inset-0 before:rounded-xl",
                                "before:bg-gradient-to-br before:from-white/20 before:to-transparent",
                                "before:opacity-50 before:pointer-events-none",
                                "after:absolute after:inset-[1px] after:rounded-[11px]",
                                "after:bg-gradient-to-br after:from-transparent after:to-black/5",
                                "after:opacity-60 after:pointer-events-none dark:after:to-white/5"
                            )}
                        >
                            <div className="flex items-center space-x-3 md:space-x-4 mb-3 md:mb-4">
                                <motion.div
                                    whileHover={{ scale: 1.05 }}
                                    transition={{ type: "spring", stiffness: 300 }}
                                >
                                    <Avatar 
                                        className="w-12 h-12 md:w-16 md:h-16 border-2 border-white/50 dark:border-white/30 shadow-lg ring-2 ring-blue-500/20 dark:ring-blue-400/20"
                                        src="https://dogeoss.grtsinry43.com/img/author.jpeg" 
                                        alt="Author"
                                        size={'4'}
                                        fallback={'Avatar'}
                                    />
                                </motion.div>
                                <div>
                                    <h2 className="text-xl md:text-2xl font-bold text-gray-800 dark:text-gray-200 drop-shadow-sm">{websiteInfo.AUTHOR_NAME}</h2>
                                    <p className="text-[0.7em] md:text-[0.75em] text-gray-600 mt-1 dark:text-gray-400">{websiteInfo.AUTHOR_INFO}</p>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <p className="text-gray-700 text-xs md:text-sm dark:text-gray-300 leading-relaxed">{websiteInfo.AUTHOR_WELCOME}</p>
                                <div className="links flex flex-col md:flex-row space-y-2 md:space-y-0 md:space-x-4">
                                    <div className="flex items-center space-x-2">
                                        <GitHubLogoIcon className="w-3 h-3 text-gray-700 dark:text-gray-300"/>
                                        <Text size="1" className="md:text-sm">
                                            <HoverCard.Root>
                                                <HoverCard.Trigger>
                                                    <a
                                                        href={websiteInfo.AUTHOR_GITHUB}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className={clsx(styles.underlineAnimation, 'text-blue-700 dark:text-blue-400', styles.glowAnimation)}
                                                    >
                                                        {websiteInfo.AUTHOR_GITHUB.replace('https://github.com/', '')}
                                                    </a>
                                                </HoverCard.Trigger>
                                                <HoverCard.Content maxWidth="300px">
                                                    <Flex gap="4">
                                                        <Avatar
                                                            size="3"
                                                            fallback="R"
                                                            radius="full"
                                                            src="https://avatars.githubusercontent.com/u/77447646?v=4"
                                                        />
                                                        <Box>
                                                            <Heading size="3" as="h3">
                                                                grtsinry43
                                                            </Heading>
                                                            <Text as="div" size="2" color="gray" mb="2">
                                                                grtsinry43 · he/him
                                                            </Text>
                                                            <Text as="div" size="2">
                                                                Nothing but enthusiasm brightens up the endless years.
                                                            </Text>
                                                        </Box>
                                                    </Flex>
                                                </HoverCard.Content>
                                            </HoverCard.Root>
                                        </Text>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        <CodeIcon className="w-3 h-3 text-gray-700 dark:text-gray-300"/>
                                        <Text size="1" className="md:text-sm">
                                            <HoverCard.Root>
                                                <HoverCard.Trigger>
                                                    <a
                                                        href="https://www.grtsinry43.com"
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className={clsx(styles.underlineAnimation, 'text-blue-700 dark:text-blue-400', styles.glowAnimation)}
                                                    >
                                                        Home Page
                                                    </a>
                                                </HoverCard.Trigger>
                                                <HoverCard.Content maxWidth="300px">
                                                    <Flex gap="4">
                                                        <Avatar
                                                            size="3"
                                                            fallback="R"
                                                            radius="full"
                                                            src="https://www.grtsinry43.com/favicon.ico"
                                                        />
                                                        <Box>
                                                            <Heading size="3" as="h3">
                                                                学习开发记录
                                                            </Heading>
                                                            <Text as="div" size="2" color="gray" mb="2">
                                                                grtsinry43 的个人主页
                                                            </Text>
                                                            <Text as="div" size="2">
                                                                记录了最近项目，学习进度，折腾历程，以及一些技术分享。
                                                            </Text>
                                                        </Box>
                                                    </Flex>
                                                </HoverCard.Content>
                                            </HoverCard.Root>
                                        </Text>
                                    </div>
                                </div>
                            </div>
                        </Container>
                    </motion.div>
                </div>
            </div>
        </motion.div>
    );
};

export default AuthorCard;
