import React, {useState, useEffect, useRef} from 'react';
import Link from 'next/link';
import {motion, AnimatePresence} from 'framer-motion';
import {Avatar, IconButton, DropdownMenu, Badge} from '@radix-ui/themes';
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

    return (
        <div className={styles.navbarWrapper} ref={navbarRef}>
            <motion.div initial={{y: -100}} animate={isInView ? {y: 0} : {y: -100}}
                        transition={{type: 'spring', stiffness: 120, damping: 20}}>
                <nav className={clsx(styles.navbar, scrolled ? styles.scrolled : '')}>
                    <div className={styles.navbarContainer}>
                        <div className={styles.avatarWrapper}>
                            <EnhancedAvatar avatarSrc={websiteInfo.WEBSITE_LOGO}/>
                        </div>
                        {isTitleVisible ? (
                            <motion.div initial={{y: 10, opacity: 0}}
                                        style={{width: '100%'}}
                                        animate={isTitleVisible ? {y: 0, opacity: 1} : {y: 10, opacity: 0}}
                                        transition={{type: 'spring', stiffness: 260, damping: 20}}>
                                <div className={clsx(article_font.className, "w-full")}
                                     style={{
                                         paddingLeft: '4rem',
                                     }}>
                                    <div className="text-[0.75em]">
                                        {/* 类型 */}
                                        <span className="font-bold mr-1">{titleInfo.type}</span>
                                        {/* 分类 */}
                                        <span>/</span>
                                        <span className="ml-1">{titleInfo.categoryName}</span>
                                    </div>
                                    {/* 标题 */}
                                    <div className={styles.title}>{titleInfo.title}</div>
                                </div>
                            </motion.div>
                        ) : (
                            <div className={styles.navItemContainer}>
                                {navItems.map((item) => (
                                    <div key={item.name} className={styles.navItemWrapper}
                                         onMouseEnter={() => handleMouseEnter(item.name)}
                                         onMouseLeave={handleMouseLeave}>
                                        <Link href={item.href} passHref>
                                            <div className={styles.navItemLink}>
                                                <motion.div whileTap={{scale: 0.95}}>
                        <span
                            className={clsx(styles.navItem, styles.underlineAnimation, styles.glowAnimation)}>{item.name}</span>
                                                </motion.div>
                                            </div>
                                        </Link>
                                        <AnimatePresence>
                                            {item.children && item.children.length > 0 && activeItem === item.name && (
                                                <motion.div initial={{opacity: 0, y: -10}} animate={{opacity: 1, y: 0}}
                                                            exit={{opacity: 0, y: -10}} transition={{duration: 0.2}}>
                                                    <div className={styles.submenu}>
                                                        {item.children.map((child, index) => (
                                                            <motion.div key={child.name}
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
                                                                        transition={{delay: index * 0.1}}>
                                                                <Link href={child.href} passHref>
                                                                    <div className={styles.submenuItemWrapper}>
                                                                        <motion.div whileTap={{scale: 0.95}}>
                                                                        <span
                                                                            className={styles.submenuItem}>{child.name}</span>
                                                                        </motion.div>
                                                                    </div>
                                                                </Link>
                                                            </motion.div>
                                                        ))}
                                                    </div>
                                                </motion.div>
                                            )}
                                        </AnimatePresence>
                                    </div>
                                ))}
                            </div>
                        )}

                        <div className={styles.navbarContent}>
                            <div className={styles.searchWrapper}>
                                <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                    <div className={styles.search}>
                                        <MagnifyingGlassIcon onClick={() => {
                                            setIsSearchVisible(true);
                                        }} className={styles.searchIcon}/>
                                    </div>
                                </motion.div>
                            </div>
                            <div className={styles.githubWrapper}>
                                <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                    <a href={websiteInfo.AUTHOR_GITHUB} target="_blank"
                                       rel="noopener noreferrer"
                                       className={styles.githubLink}>
                                        <GitHubLogoIcon/>
                                    </a>
                                </motion.div>
                            </div>
                            <div className={styles.themeToggleWrapper}>
                                <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                    {mounted && (
                                        <div onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
                                             className={styles.themeToggle}>
                                            {resolvedTheme === 'dark' ? <SunIcon/> : <MoonIcon/>}
                                        </div>
                                    )}
                                </motion.div>
                            </div>
                            <div className={styles.loginButtonWrapper}>
                                {user.isLogin ? (
                                    <div className={styles.avatarWrapper}>
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

                                                {/*<DropdownMenu.Sub>*/}
                                                {/*    <DropdownMenu.SubTrigger> 更多操作 </DropdownMenu.SubTrigger>*/}
                                                {/*    <DropdownMenu.SubContent>*/}
                                                {/*        <DropdownMenu.Item> 设置 </DropdownMenu.Item>*/}
                                                {/*        /!*<DropdownMenu.Item> 我的收藏 </DropdownMenu.Item>*!/*/}
                                                {/*    </DropdownMenu.SubContent>*/}
                                                {/*</DropdownMenu.Sub>*/}

                                                <DropdownMenu.Separator/>
                                                <DropdownMenu.Item color="red" onClick={() => {
                                                    dispatch({type: 'user/clearUserInfo', payload: null});
                                                    dispatch({type: 'user/changeLoginStatus', payload: false});
                                                }}>
                                                    退出登录
                                                </DropdownMenu.Item>
                                            </DropdownMenu.Content>
                                        </DropdownMenu.Root>
                                    </div>
                                ) : (
                                    <motion.div whileHover={{scale: 1.05}} whileTap={{scale: 0.95}}>
                                        <IconButton variant="ghost" radius={'full'} color={'gray'}
                                                    className={styles.loginButton}
                                                    onClick={openLoginModal}>
                                            <UserRoundPlusIcon width={16} height={16}/>
                                        </IconButton>
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
