"use client"

import {useEffect, useState} from "react"
import {useTheme} from "next-themes"
import {motion, AnimatePresence} from "framer-motion"
import Link from "next/link"
import {Avatar} from "@/components/ui/avatar"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {Button} from "@/components/ui/button"
import {Github, Menu, Moon, Search, Sun, UserPlus, LogOut, Settings} from "lucide-react"
import {cn} from "@/lib/utils"
import {useRouter} from "next/navigation"

// Assuming these imports exist in your project
import emitter from "@/utils/eventBus"
import type {TitleEvent} from "@/components/article/ArticleScrollSync"
import {article_font} from "@/app/fonts/font"
import {useWebsiteInfo} from "@/app/website-info-provider"
import LoginModalMobile from "@/components/user/LoginModalMobile"
import {useAppDispatch, useAppSelector} from "@/redux/hooks"
import SearchModal from "@/components/SearchModal"
import {EnhancedAvatar} from "@/components/navbar/EnhancedAvatar"
import {Badge} from "@/components/ui/badge"
import styles from "@/styles/NavBarMobile.module.scss"

type NavItem = {
    name: string
    href: string
    children?: { name: string; href: string }[]
}

const NavBarMobile = ({items}: { items: NavItem[] }) => {
    const {resolvedTheme, setTheme} = useTheme()
    const [mounted, setMounted] = useState(false)
    const [showPanel, setShowPanel] = useState(false)
    const [isTitleVisible, setIsTitleVisible] = useState(false)
    const [titleInfo, setTitleInfo] = useState({title: "", categoryName: "", type: ""})
    const [scrolled, setScrolled] = useState(false)
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false)
    const [isSearchVisible, setIsSearchVisible] = useState(false)

    const websiteInfo = useWebsiteInfo()
    const router = useRouter()
    const dispatch = useAppDispatch()
    const user = useAppSelector((state) => state.user)

    // Mount effect
    useEffect(() => {
        setMounted(true)
    }, [])

    // Title visibility effect
    useEffect(() => {
        const showTitleHandler = (config: TitleEvent) => {
            setIsTitleVisible(true)
            setTitleInfo({...config})
        }

        const hideTitleHandler = () => {
            setIsTitleVisible(false)
        }

        // @ts-expect-error - Event typing issue
        emitter.on("showTitle", showTitleHandler)
        emitter.on("hideTitle", hideTitleHandler)

        return () => {
            // @ts-expect-error - Event typing issue
            emitter.off("showTitle", showTitleHandler)
            emitter.off("hideTitle", hideTitleHandler)
        }
    }, [])

    // Scroll effect
    useEffect(() => {
        const handleScroll = () => {
            setScrolled(window.scrollY > 20)
        }

        window.addEventListener("scroll", handleScroll)
        return () => window.removeEventListener("scroll", handleScroll)
    }, [])

    const togglePanel = () => {
        setShowPanel(!showPanel)
    }

    const handleLogout = () => {
        dispatch({type: "user/clearUserInfo", payload: null})
        dispatch({type: "user/changeLoginStatus", payload: false})
    }

    // 优化的动画配置
    const springConfig = {
        type: 'spring' as const,
        stiffness: 400,
        damping: 28,
        mass: 0.7
    }

    const smoothTransition = {
        type: 'spring' as const,
        stiffness: 300,
        damping: 22
    }

    const quickTransition = {
        duration: 0.2,
        ease: "easeOut" as const
    }

    return (
        <div className={styles.navbarWrapper}>
            <motion.div 
                initial={{y: -100}} 
                animate={{y: 0}} 
                transition={springConfig}
            >
                <nav
                    className={cn(
                        styles.navbar,
                        scrolled ? styles.scrolled : ""
                    )}
                >
                    <div className={styles.navbarContainer}>
                        <motion.div
                            whileHover={{ scale: 1.02 }}
                            whileTap={{ scale: 0.98 }}
                            transition={quickTransition}
                        >
                            <Button 
                                variant="ghost" 
                                size="icon" 
                                onClick={() => {
                                    togglePanel()
                                    setIsTitleVisible(false)
                                }} 
                                className={styles.menuButton}
                            >
                                <Menu className="h-5 w-5"/>
                            </Button>
                        </motion.div>

                        <AnimatePresence mode="wait">
                            {isTitleVisible ? (
                                <motion.div
                                    key="title"
                                    initial={{y: 20, opacity: 0}}
                                    animate={{y: 0, opacity: 1}}
                                    exit={{y: -20, opacity: 0}}
                                    transition={smoothTransition}
                                    className={styles.titleArea}
                                >
                                    <div className={cn(article_font.className)}>
                                        <motion.div 
                                            className={styles.titleType}
                                            initial={{x: -10, opacity: 0}}
                                            animate={{x: 0, opacity: 1}}
                                            transition={{...smoothTransition, delay: 0.1}}
                                        >
                                            <span className="font-bold mr-1">{titleInfo.type}</span>
                                            <span>/</span>
                                            <span className="ml-1">{titleInfo.categoryName}</span>
                                        </motion.div>
                                        <motion.div 
                                            className={styles.titleText}
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
                                    key="navigation"
                                    initial={{y: -20, opacity: 0}}
                                    animate={{y: 0, opacity: 1}}
                                    exit={{y: 20, opacity: 0}}
                                    transition={smoothTransition}
                                    className={styles.buttonGroup}
                                >
                                    <motion.div 
                                        className={styles.avatarWrapper}
                                        whileHover={{ scale: 1.02 }}
                                        whileTap={{ scale: 0.98 }}
                                        transition={smoothTransition}
                                    >
                                        <EnhancedAvatar avatarSrc={websiteInfo.WEBSITE_LOGO}/>
                                    </motion.div>

                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <motion.div 
                                            whileHover={{scale: 1.05}} 
                                            whileTap={{scale: 0.95}}
                                            transition={quickTransition}
                                        >
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => setIsSearchVisible(true)}
                                                className={styles.iconButton}
                                            >
                                                <Search className="h-[18px] w-[18px]"/>
                                            </Button>
                                        </motion.div>

                                        <motion.div 
                                            whileHover={{scale: 1.05, rotate: 3}} 
                                            whileTap={{scale: 0.95}}
                                            transition={quickTransition}
                                        >
                                            <Button variant="ghost" size="icon" asChild
                                                    className={styles.iconButton}>
                                                <a href={websiteInfo.AUTHOR_GITHUB} target="_blank"
                                                   rel="noopener noreferrer">
                                                    <Github className="h-[18px] w-[18px]"/>
                                                </a>
                                            </Button>
                                        </motion.div>

                                        {mounted && (
                                            <motion.div 
                                                whileHover={{scale: 1.05}} 
                                                whileTap={{scale: 0.95}}
                                                transition={quickTransition}
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
                                                    className={styles.iconButton}
                                                >
                                                    <AnimatePresence mode="wait">
                                                        <motion.div
                                                            key={resolvedTheme}
                                                            initial={{ scale: 0, rotate: -90 }}
                                                            animate={{ scale: 1, rotate: 0 }}
                                                            exit={{ scale: 0, rotate: 90 }}
                                                            transition={{ duration: 0.2 }}
                                                        >
                                                            {resolvedTheme === "dark" ? (
                                                                <Sun className="h-[18px] w-[18px]"/>
                                                            ) : (
                                                                <Moon className="h-[18px] w-[18px]"/>
                                                            )}
                                                        </motion.div>
                                                    </AnimatePresence>
                                                </Button>
                                            </motion.div>
                                        )}

                                        {user.isLogin ? (
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <motion.div
                                                        whileHover={{ scale: 1.05 }}
                                                        whileTap={{ scale: 0.95 }}
                                                        transition={quickTransition}
                                                    >
                                                        <Button variant="ghost" size="icon"
                                                                className={cn(styles.iconButton, "rounded-full overflow-hidden")}>
                                                            <Avatar className="h-8 w-8">
                                                                <img src={user.userInfo.avatar || undefined}
                                                                     alt={user.userInfo.nickname || "User"}/>
                                                                <span
                                                                    className="sr-only">{user.userInfo.nickname || "User"}</span>
                                                            </Avatar>
                                                        </Button>
                                                    </motion.div>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end" className={styles.userDropdown}>
                                                    <div className="flex items-center justify-start gap-2 p-2">
                                                        <Avatar className="h-8 w-8">
                                                            <img src={user.userInfo.avatar || undefined}
                                                                 alt={user.userInfo.nickname || "User"}/>
                                                        </Avatar>
                                                        <div className="flex flex-col">
                                                            <p className="text-sm font-medium leading-none">{user.userInfo.nickname}</p>
                                                            <p className="text-xs text-muted-foreground">{user.userInfo.email}</p>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center px-2 py-1">
                                                        <Badge variant="outline" className="text-xs">
                                                            {user.userInfo.oauthProvider || "本站"}
                                                        </Badge>
                                                    </div>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem 
                                                        onClick={() => router.push("/my")}
                                                        className={styles.dropdownItem}
                                                    >
                                                        <Settings className="mr-2 h-4 w-4"/>
                                                        <span>用户中心与设置</span>
                                                    </DropdownMenuItem>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem 
                                                        onClick={handleLogout}
                                                        className={cn(styles.dropdownItem, styles.destructive)}
                                                    >
                                                        <LogOut className="mr-2 h-4 w-4"/>
                                                        <span>退出登录</span>
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        ) : (
                                            <motion.div 
                                                whileHover={{scale: 1.03}} 
                                                whileTap={{scale: 0.97}}
                                                transition={quickTransition}
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    onClick={() => setIsLoginModalOpen(true)}
                                                    className={styles.iconButton}
                                                >
                                                    <motion.div
                                                        whileHover={{ rotate: 8 }}
                                                        transition={{ duration: 0.15 }}
                                                    >
                                                        <UserPlus className="h-[18px] w-[18px]"/>
                                                    </motion.div>
                                                </Button>
                                            </motion.div>
                                        )}
                                    </div>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </div>
                </nav>
            </motion.div>

            {/* Navigation Panel */}
            <AnimatePresence>
                {showPanel && (
                    <motion.div
                        initial={{opacity: 0, height: 0, y: -20}}
                        animate={{opacity: 1, height: "auto", y: 0}}
                        exit={{opacity: 0, height: 0, y: -20}}
                        transition={{duration: 0.25, ease: "easeInOut"}}
                        className={styles.navigationPanel}
                        style={{ top: '60px' }}
                    >
                        <div className={styles.panelContent}>
                            {items.map((item, index) => (
                                <motion.div
                                    key={item.name}
                                    initial={{opacity: 0, y: -10, x: -20}}
                                    animate={{opacity: 1, y: 0, x: 0}}
                                    transition={{
                                        ...smoothTransition,
                                        delay: index * 0.04
                                    }}
                                    style={{ marginBottom: '8px' }}
                                >
                                    <Link
                                        href={item.href}
                                        className={styles.navItem}
                                        onClick={() => setShowPanel(false)}
                                    >
                                        {item.name}
                                    </Link>

                                    {item.children && item.children.length > 0 && (
                                        <div className={styles.subNavContainer}>
                                            {item.children.map((child, childIndex) => (
                                                <motion.div
                                                    key={child.name}
                                                    initial={{opacity: 0, x: -15}}
                                                    animate={{opacity: 1, x: 0}}
                                                    transition={{
                                                        ...smoothTransition,
                                                        delay: index * 0.04 + childIndex * 0.02 + 0.1
                                                    }}
                                                    style={{ marginBottom: '4px' }}
                                                >
                                                    <Link
                                                        href={child.href}
                                                        className={styles.subNavItem}
                                                        onClick={() => setShowPanel(false)}
                                                    >
                                                        {child.name}
                                                    </Link>
                                                </motion.div>
                                            ))}
                                        </div>
                                    )}
                                </motion.div>
                            ))}
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* Background gradient */}
            <motion.div
                initial={{opacity: 0}}
                animate={{opacity: 1}}
                transition={{duration: 0.5}}
                className={styles.background}
            />

            {/* Modals */}
            <LoginModalMobile isOpen={isLoginModalOpen} onClose={() => setIsLoginModalOpen(false)}/>
            <SearchModal open={isSearchVisible} setOpen={setIsSearchVisible}/>
        </div>
    )
}

export default NavBarMobile

