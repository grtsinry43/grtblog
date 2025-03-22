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
import {Badge} from "@/components/ui/badge";

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

    return (
        <div className="fixed w-full z-50">
            <motion.div initial={{y: -100}} animate={{y: 0}} transition={{type: "spring", stiffness: 120, damping: 20}}>
                <nav
                    className={cn(
                        "py-3 px-4 transition-all duration-300",
                        scrolled ? "bg-background/95 backdrop-blur-md shadow-md" : "bg-transparent",
                    )}
                >
                    <div className="flex items-center justify-between">
                        <Button variant="ghost" size="icon" onClick={()=>{
                            togglePanel()
                            setIsTitleVisible(false)
                        }} className="relative z-50">
                            <Menu className="h-5 w-5"/>
                        </Button>

                        {isTitleVisible ? (
                            <motion.div
                                initial={{y: 10, opacity: 0}}
                                animate={{y: 0, opacity: 1}}
                                transition={{type: "spring", stiffness: 260, damping: 20}}
                                className="flex-1 pl-4"
                            >
                                <div className={cn(article_font.className, "w-full")}>
                                    <div className="text-xs text-muted-foreground">
                                        <span className="font-bold mr-1">{titleInfo.type}</span>
                                        <span>/</span>
                                        <span className="ml-1">{titleInfo.categoryName}</span>
                                    </div>
                                    <div className="text-sm font-medium truncate">{titleInfo.title}</div>
                                </div>
                            </motion.div>
                        ) : (
                            <>
                                <div className="flex-shrink-0">
                                    <EnhancedAvatar avatarSrc={websiteInfo.WEBSITE_LOGO}/>
                                </div>

                                <div className="flex items-center gap-2">
                                    <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => setIsSearchVisible(true)}
                                            className="text-foreground/80 hover:text-foreground"
                                        >
                                            <Search className="h-[18px] w-[18px]"/>
                                        </Button>
                                    </motion.div>

                                    <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                        <Button variant="ghost" size="icon" asChild
                                                className="text-foreground/80 hover:text-foreground">
                                            <a href={websiteInfo.AUTHOR_GITHUB} target="_blank"
                                               rel="noopener noreferrer">
                                                <Github className="h-[18px] w-[18px]"/>
                                            </a>
                                        </Button>
                                    </motion.div>

                                    {mounted && (
                                        <motion.div whileHover={{scale: 1.1}} whileTap={{scale: 0.95}}>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
                                                className="text-foreground/80 hover:text-foreground"
                                            >
                                                {resolvedTheme === "dark" ? (
                                                    <Sun className="h-[18px] w-[18px]"/>
                                                ) : (
                                                    <Moon className="h-[18px] w-[18px]"/>
                                                )}
                                            </Button>
                                        </motion.div>
                                    )}

                                    {user.isLogin ? (
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" size="icon"
                                                        className="rounded-full overflow-hidden">
                                                    <Avatar className="h-8 w-8">
                                                        <img src={user.userInfo.avatar || undefined}
                                                             alt={user.userInfo.nickname || "User"}/>
                                                        <span
                                                            className="sr-only">{user.userInfo.nickname || "User"}</span>
                                                    </Avatar>
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end" className="w-56">
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
                                                <DropdownMenuItem onClick={() => router.push("/my")}>
                                                    <Settings className="mr-2 h-4 w-4"/>
                                                    <span>用户中心与设置</span>
                                                </DropdownMenuItem>
                                                <DropdownMenuSeparator/>
                                                <DropdownMenuItem onClick={handleLogout}
                                                                  className="text-destructive focus:text-destructive">
                                                    <LogOut className="mr-2 h-4 w-4"/>
                                                    <span>退出登录</span>
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    ) : (
                                        <motion.div whileHover={{scale: 1.05}} whileTap={{scale: 0.95}}>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => setIsLoginModalOpen(true)}
                                                className="text-foreground/80 hover:text-foreground"
                                            >
                                                <UserPlus className="h-[18px] w-[18px]"/>
                                            </Button>
                                        </motion.div>
                                    )}
                                </div>
                            </>
                        )}
                    </div>
                </nav>
            </motion.div>

            {/* Navigation Panel */}
            <AnimatePresence>
                {showPanel && (
                    <motion.div
                        initial={{opacity: 0, height: 0}}
                        animate={{opacity: 1, height: "auto"}}
                        exit={{opacity: 0, height: 0}}
                        transition={{duration: 0.3, ease: "easeInOut"}}
                        className="fixed inset-x-0 top-[60px] bg-background/95 backdrop-blur-md shadow-lg rounded-b-xl z-40"
                    >
                        <div className="max-h-[70vh] overflow-y-auto py-4 px-6">
                            {items.map((item, index) => (
                                <motion.div
                                    key={item.name}
                                    initial={{opacity: 0, y: -10}}
                                    animate={{opacity: 1, y: 0}}
                                    transition={{delay: index * 0.05}}
                                    className="py-2"
                                >
                                    <Link
                                        href={item.href}
                                        className="block text-lg font-medium text-foreground hover:text-primary transition-colors duration-200"
                                        onClick={() => setShowPanel(false)}
                                    >
                                        {item.name}
                                    </Link>

                                    {item.children && item.children.length > 0 && (
                                        <div className="mt-2 ml-4 space-y-1 border-l-2 border-muted pl-3">
                                            {item.children.map((child, childIndex) => (
                                                <motion.div
                                                    key={child.name}
                                                    initial={{opacity: 0, x: -5}}
                                                    animate={{opacity: 1, x: 0}}
                                                    transition={{delay: index * 0.05 + childIndex * 0.03 + 0.1}}
                                                >
                                                    <Link
                                                        href={child.href}
                                                        className="block py-1 text-sm text-muted-foreground hover:text-primary transition-colors duration-200"
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
                className="fixed inset-0 bg-gradient-to-b from-primary/5 via-background/0 to-background/0 -z-10 pointer-events-none"
            />

            {/* Modals */}
            <LoginModalMobile isOpen={isLoginModalOpen} onClose={() => setIsLoginModalOpen(false)}/>
            <SearchModal open={isSearchVisible} setOpen={setIsSearchVisible}/>
        </div>
    )
}

export default NavBarMobile

