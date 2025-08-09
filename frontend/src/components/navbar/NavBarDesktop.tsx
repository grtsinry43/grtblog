import React, {useState, useEffect, useRef} from 'react';
import Link from 'next/link';
import {motion, AnimatePresence} from 'framer-motion';
import {Avatar, DropdownMenu, Badge} from '@radix-ui/themes';
import {Button} from '@/components/ui/button';
import {MoonIcon, SunIcon, GitHubLogoIcon, MagnifyingGlassIcon} from '@radix-ui/react-icons';
import styles from '@/styles/NavBar.module.scss';
import {useTheme} from 'next-themes';
import {clsx} from 'clsx';
import LoginModal from '@/components/user/LoginModal';
import {useAppDispatch, useAppSelector} from '@/redux/hooks';
import {User} from '@/redux/userSlice';
import {userInfo} from '@/api/user';
import {UserRoundPlusIcon} from 'lucide-react';
import emitter from "@/utils/eventBus";
import {TitleEvent} from "@/components/article/ArticleScrollSync";
import {article_font} from "@/app/fonts/font";
import {OnlineCount} from "@/redux/onlineCountSlice";
import {useWebsiteInfo} from "@/app/website-info-provider";
import {useRouter} from "next/navigation";
import SearchModal from "@/components/SearchModal";
import {EnhancedAvatar} from "@/components/navbar/EnhancedAvatar";

export default function NavBarDesktop({items}: {
    items: { name: string; href: string; children?: { name: string; href: string }[] }[]
}) {
    const {resolvedTheme, setTheme, theme} = useTheme();
    const [activeItem, setActiveItem] = useState<string | null>(null);
    const [mounted, setMounted] = useState(false);
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);

    const user = useAppSelector((state: { user: User, onlineCount: OnlineCount }) => state.user);
    const dispatch = useAppDispatch();

    const [isTitleVisible, setIsTitleVisible] = useState(false);
    const [titleInfo, setTitleInfo] = useState({title: '', categoryName: '', type: ''});

    const navbarRef = useRef<HTMLDivElement>(null);
    const [isInView, setIsInView] = useState(false);
    const websiteInfo = useWebsiteInfo();
    const router = useRouter();

    const [isSearchVisible, setIsSearchVisible] = useState(false);

    useEffect(() => {
        setMounted(true);
        userInfo().then((res) => {
            if (res) {
                dispatch({type: 'user/initUserInfo', payload: res});
                dispatch({type: 'user/changeLoginStatus', payload: true});
            }
        });
    }, [dispatch]);

    useEffect(() => {
        document.documentElement.classList.add('transition-colors');
        return () => {
            document.documentElement.classList.remove('transition-colors');
        };
    }, [theme]);

    useEffect(() => {
        const showTitleHandler = (config: TitleEvent) => {
            setIsTitleVisible(true);
            setTitleInfo({...config});
        };

        const hideTitleHandler = () => {
            setIsTitleVisible(false);
        }

        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        emitter.on('showTitle', showTitleHandler);
        emitter.on('hideTitle', hideTitleHandler);

        return () => {
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-expect-error
            emitter.off('showTitle', showTitleHandler);
            emitter.off('hideTitle', hideTitleHandler);
        };
    }, []);

    useEffect(() => {
        setIsTitleVisible(false);
    }, []);

    useEffect(() => {
        const observer = new IntersectionObserver(
            ([entry]) => {
                setIsInView(entry.isIntersecting);
            },
            {threshold: 0.1}
        );

        if (navbarRef.current) {
            observer.observe(navbarRef.current);
        }

        return () => {
            if (navbarRef.current) {
                observer.unobserve(navbarRef.current);
            }
        };
    }, []);

    const navItems = items;

    const handleMouseEnter = (name: string) => {
        setActiveItem(name);
    };

    const handleMouseLeave = () => {
        setActiveItem(null);
    };

    const openLoginModal = () => {
        setIsLoginModalOpen(true);
    };

    const closeLoginModal = () => {

        setIsLoginModalOpen(false);
    };

    const [scrolled, setScrolled] = useState(false);

    useEffect(() => {
        const handleScroll = () => {
            setScrolled(window.scrollY > 20);
        };

        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    // 优化的动画配置 - 更快的响应
    const springConfig = {
        type: 'spring' as const,
        stiffness: 400,  // 增加刚度，更快响应
        damping: 28,     // 调整阻尼
        mass: 0.7        // 减小质量，更轻快
    };

    const smoothTransition = {
        type: 'spring' as const,
        stiffness: 300,  // 增加刚度
        damping: 22      // 调整阻尼
    };

    // 更快的简单过渡
    const quickTransition = {
        duration: 0.2,
        ease: "easeOut" as const
    };

    return (
        <div className={styles.navbarWrapper} ref={navbarRef}>
            <motion.div 
                initial={{y: -100}} 
                animate={isInView ? {y: 0} : {y: -100}}
                transition={springConfig}
            >
                <nav className={clsx(styles.navbar, scrolled ? styles.scrolled : '')}>
                    <div className={styles.navbarContainer}>
                        <motion.div 
                            className={styles.avatarWrapper}
                            whileHover={{ scale: 1.02 }}
                            whileTap={{ scale: 0.98 }}
                            transition={smoothTransition}
                        >
                            <EnhancedAvatar avatarSrc={websiteInfo.WEBSITE_LOGO}/>
                        </motion.div>
                        
                        <AnimatePresence mode="wait">
                            {isTitleVisible ? (
                                <motion.div 
                                    key="title"
                                    initial={{y: 20, opacity: 0}}
                                    animate={{y: 0, opacity: 1}}
                                    exit={{y: -20, opacity: 0}}
                                    transition={smoothTransition}
                                    style={{width: '100%'}}
                                >
                                    <div className={clsx(article_font.className, "w-full")}
                                         style={{
                                             paddingLeft: '4rem',
                                         }}>
                                        <motion.div 
                                            className="text-[0.75em]"
                                            initial={{x: -10, opacity: 0}}
                                            animate={{x: 0, opacity: 1}}
                                            transition={{...smoothTransition, delay: 0.1}}
                                        >
                                            {/* 类型 */}
                                            <span className="font-bold mr-1">{titleInfo.type}</span>
                                            {/* 分类 */}
                                            <span>/</span>
                                            <span className="ml-1">{titleInfo.categoryName}</span>
                                        </motion.div>
                                        {/* 标题 */}
                                        <motion.div 
                                            style={{
                                                lineHeight: '1.2em',
                                            }}
                                            initial={{x: -10, opacity: 0}}
                                            animate={{x: 0, opacity: 1}}
                                            transition={{...smoothTransition, delay: 0.2}}
                                        >
                                            {titleInfo.title}
                                        </motion.div>
                                    </div>
                                </motion.div>
                            ) : (
                                <motion.div
                                    key="nav"
                                    initial={{y: -20, opacity: 0}}
                                    animate={{y: 0, opacity: 1}}
                                    exit={{y: 20, opacity: 0}}
                                    transition={smoothTransition}
                                    style={{width: '100%'}}
                                >
                                    <div className={styles.navItemContainer}>
                                        {navItems.map((item, index) => (
                                            <motion.div 
                                                key={item.name} 
                                                className={styles.navItemWrapper}
                                                onMouseEnter={() => handleMouseEnter(item.name)}
                                                onMouseLeave={handleMouseLeave}
                                                initial={{y: -20, opacity: 0}}
                                                animate={{y: 0, opacity: 1}}
                                                transition={{...smoothTransition, delay: index * 0.05}}
                                                whileHover={{ 
                                                    y: -1,  // 更小的浮动距离
                                                    transition: { duration: 0.15 }  // 更快的响应
                                                }}
                                            >
                                                <Link href={item.href} passHref>
                                                    <div className={styles.navItemLink}>
                                                        <motion.div whileTap={{scale: 0.95}}>
                                                            <span
                                                                className={clsx(styles.navItem, styles.underlineAnimation, styles.glowAnimation)}
                                                            >
                                                                {item.name}
                                                            </span>
                                                        </motion.div>
                                                    </div>
                                                </Link>
                                                <AnimatePresence>
                                                    {item.children && item.children.length > 0 && activeItem === item.name && (
                                                        <motion.div 
                                                            initial={{opacity: 0, y: -15, scale: 0.95}}
                                                            animate={{opacity: 1, y: 0, scale: 1}}
                                                            exit={{opacity: 0, y: -15, scale: 0.95}}
                                                            transition={smoothTransition}
                                                        >
                                                            <div className={styles.submenu}>
                                                                {item.children.map((child, index) => (
                                                                    <motion.div 
                                                                        key={child.name}
                                                                        initial={{
                                                                            opacity: 0,
                                                                            x: -20,
                                                                            filter: 'blur(10px)'
                                                                        }}
                                                                        animate={{
                                                                            opacity: 1,
                                                                            x: 0,
                                                                            filter: 'blur(0px)'
                                                                        }}
                                                                        transition={{
                                                                            ...smoothTransition,
                                                                            delay: index * 0.05  // 更快的出现延迟
                                                                        }}
                                                                    >
                                                                        <Link href={child.href} passHref>
                                                                            <div className={styles.submenuItemWrapper}>
                                                                                <motion.div whileTap={{scale: 0.95}}>
                                                                                    <span className={styles.submenuItem}>
                                                                                        {child.name}
                                                                                    </span>
                                                                                </motion.div>
                                                                            </div>
                                                                        </Link>
                                                                    </motion.div>
                                                                ))}
                                                            </div>
                                                        </motion.div>
                                                    )}
                                                </AnimatePresence>
                                            </motion.div>
                                        ))}
                                    </div>
                                </motion.div>
                            )}
                        </AnimatePresence>

                        <div className={styles.navbarContent}>
                            <div className={styles.searchWrapper}>
                                <motion.div 
                                    whileHover={{scale: 1.05}}  // 更小的缩放
                                    whileTap={{scale: 0.95}}
                                    transition={quickTransition}
                                >
                                    <div className={styles.search}>
                                        <MagnifyingGlassIcon 
                                            onClick={() => {
                                                setIsSearchVisible(true);
                                            }} 
                                            className={styles.searchIcon}
                                        />
                                    </div>
                                </motion.div>
                            </div>
                            <div className={styles.githubWrapper}>
                                <motion.div 
                                    whileHover={{scale: 1.05, rotate: 3}}  // 更小的缩放和旋转
                                    whileTap={{scale: 0.95}}
                                    transition={quickTransition}
                                >
                                    <a href={websiteInfo.AUTHOR_GITHUB} target="_blank"
                                       rel="noopener noreferrer"
                                       className={styles.githubLink}>
                                        <GitHubLogoIcon/>
                                    </a>
                                </motion.div>
                            </div>
                            <div className={styles.themeToggleWrapper}>
                                <motion.div 
                                    whileHover={{scale: 1.05}}  // 更小的缩放
                                    whileTap={{scale: 0.95}}
                                    transition={quickTransition}
                                >
                                    {mounted && (
                                        <motion.div 
                                            onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
                                            className={styles.themeToggle}
                                            animate={{ rotate: resolvedTheme === 'dark' ? 180 : 0 }}
                                            transition={{ duration: 0.3, ease: "easeInOut" }}  // 更快的主题切换
                                        >
                                            <AnimatePresence mode="wait">
                                                <motion.div
                                                    key={resolvedTheme}
                                                    initial={{ scale: 0, rotate: -90 }}  // 更小的旋转角度
                                                    animate={{ scale: 1, rotate: 0 }}
                                                    exit={{ scale: 0, rotate: 90 }}
                                                    transition={{ duration: 0.2 }}  // 更快的图标切换
                                                >
                                                    {resolvedTheme === 'dark' ? <SunIcon/> : <MoonIcon/>}
                                                </motion.div>
                                            </AnimatePresence>
                                        </motion.div>
                                    )}
                                </motion.div>
                            </div>
                            <div className={styles.loginButtonWrapper}>
                                {user.isLogin ? (
                                    <motion.div 
                                        className={styles.avatarWrapper}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                        transition={smoothTransition}
                                    >
                                        <DropdownMenu.Root>
                                            <DropdownMenu.Trigger>
                                                <Avatar
                                                    size="3"
                                                    radius="large"
                                                    src={user.userInfo.avatar ? user.userInfo.avatar : undefined}
                                                    fallback={user.userInfo.nickname ? user.userInfo.nickname[0].toUpperCase() : 'U'}
                                                    className={styles.avatar}
                                                />
                                            </DropdownMenu.Trigger>
                                            <DropdownMenu.Content>
                                                <DropdownMenu.Item>{user.userInfo.nickname}
                                                    <Badge className={styles.tag}
                                                           color="gray">{user.userInfo.oauthProvider ? user.userInfo.oauthProvider : '本站'}</Badge>
                                                </DropdownMenu.Item>
                                                <DropdownMenu.Item>{user.userInfo.email}</DropdownMenu.Item>
                                                <DropdownMenu.Separator/>
                                                <DropdownMenu.Item onClick={() => {
                                                    router.push('/my')
                                                }}> 用户中心与设置 </DropdownMenu.Item>

                                                <DropdownMenu.Separator/>
                                                <DropdownMenu.Item color="red" onClick={() => {
                                                    dispatch({type: 'user/clearUserInfo', payload: null});
                                                    dispatch({type: 'user/changeLoginStatus', payload: false});
                                                }}>
                                                    退出登录
                                                </DropdownMenu.Item>
                                            </DropdownMenu.Content>
                                        </DropdownMenu.Root>
                                    </motion.div>
                                ) : (
                                    <motion.div 
                                        whileHover={{scale: 1.03}}  // 更小的缩放
                                        whileTap={{scale: 0.97}}
                                        transition={quickTransition}
                                    >
                                        <Button variant="ghost"
                                                style={{
                                                    width: '2.3em',
                                                    height: '2.3em',
                                                    overflow: 'hidden',
                                                    borderRadius: '50%',
                                                }}
                                                className={styles.loginButton}
                                                onClick={openLoginModal}>
                                            <motion.div
                                                whileHover={{ rotate: 8 }}  // 更小的旋转角度
                                                transition={{ duration: 0.15 }}  // 更快的响应
                                            >
                                                <UserRoundPlusIcon width={16} height={16}/>
                                            </motion.div>
                                        </Button>
                                    </motion.div>
                                )}
                            </div>
                        </div>
                    </div>
                </nav>
            </motion.div>
            <LoginModal isOpen={isLoginModalOpen} onClose={closeLoginModal}/>
            <SearchModal open={isSearchVisible} setOpen={setIsSearchVisible}/>
        </div>
    )
        ;
}
