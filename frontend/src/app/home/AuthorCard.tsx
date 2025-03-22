'use client';

import React from 'react';
import styles from '@/styles/Home.module.scss';
import {Avatar, Container, HoverCard, Text, Flex, Box, Heading} from '@radix-ui/themes';
import {CodeIcon, GitHubLogoIcon} from '@radix-ui/react-icons';
import {clsx} from 'clsx';
import {motion} from 'framer-motion';
import {useWebsiteInfo} from "@/app/website-info-provider";

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
                    <div
                        className={clsx('absolute -left-16 -bottom-16 w-32 h-32 bg-blue-500 rounded-full blur-2xl opacity-50 animate-pulse', styles.bannerCardDotLeft)}/>
                    <div
                        className={clsx('absolute -right-16 -top-16 w-32 h-32 bg-orange-500 rounded-full blur-2xl opacity-50 animate-pulse', styles.bannerCardDotRight)}/>
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
                    >
                        <Container
                            className="z-30 relative backdrop-blur-3xl bg-gradient-to-r from-white/40 to-white/20 dark:from-black/40 dark:to-black/20 p-8 rounded-xl border border-white/20 shadow-2xl">
                            <div className="flex items-center space-x-4 mb-4">
                                <Avatar className="w-16 h-16 border-2 border-white shadow-lg"
                                        src="https://dogeoss.grtsinry43.com/img/author.jpeg" alt="Author"
                                        size={'4'}
                                        fallback={'Avatar'}/>
                                <div>
                                    <h2 className="text-2xl font-bold text-gray-800 dark:text-gray-300">{websiteInfo.AUTHOR_NAME}</h2>
                                    <p className="text-sm text-gray-600 mt-1 dark:text-gray-400">{websiteInfo.AUTHOR_INFO}</p>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <p
                                    className="text-gray-700 dark:text-gray-300">{websiteInfo.AUTHOR_WELCOME}</p>
                                <div className="links flex">
                                    <div className="flex items-center space-x-2">
                                        <GitHubLogoIcon className="w-5 h-5 text-gray-700 dark:text-gray-300"/>
                                        <Text>
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
                                    <div className="flex items-center space-x-2 ml-4">
                                        <CodeIcon className="w-5 h-5 text-gray-700 dark:text-gray-300"/>
                                        <Text>
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
