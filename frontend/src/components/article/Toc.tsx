"use client"

import {useEffect, useState, useRef, useCallback} from "react"
import {motion, AnimatePresence} from "framer-motion"
import {useTheme} from "next-themes"
import useIsMobile from "@/hooks/useIsMobile"
import emitter from "@/utils/eventBus"
import CommentModal from "@/components/comment/CommentModal"
import {Button} from "@/components/ui/button"
import ReadingProgress from "@/components/article/ReadingProgress"
import {clsx} from "clsx"
import {article_font} from "@/app/fonts/font"
import {likeRequest} from "@/api/like"
import {toast} from "react-toastify"
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from "@/components/ui/tooltip"
import {Heart, MessageCircle, Share2, BookOpen} from "lucide-react"

export type TocItem = {
    level: number
    name: string
    isSelect?: boolean
    anchor: string
    children?: TocItem[]
}

export default function Toc({
                                toc,
                                commentId,
                                targetId,
                                likes,
                                comments,
                                type,
                            }: {
    toc: TocItem[]
    commentId: string
    targetId: string
    likes: number
    comments: number
    type: "article" | "moment" | "page"
}) {
    const isMobile = useIsMobile()
    const [activeAnchor, setActiveAnchor] = useState<string | null>(null)
    const {theme} = useTheme()
    const isDark = theme === "dark"
    const [likesNum, setLikesNum] = useState(likes)
    const tocRef = useRef<HTMLDivElement>(null)
    const [activeItemRef, setActiveItemRef] = useState<HTMLLIElement | null>(null)
    const [isCommentOpen, setIsCommentOpen] = useState(false)
    const [liked, setLiked] = useState(false)
    const [isCollapsed, setIsCollapsed] = useState(false)

    const spring = {
        type: "spring",
        stiffness: 300,
        damping: 30,
        mass: 0.2,
        bounce: 0.5,
    }

    const containerVariants = {
        hidden: {opacity: 0},
        visible: {
            opacity: 1,
            transition: {
                staggerChildren: 0.05,
            },
        },
    }

    const itemVariants = {
        hidden: {x: -20, opacity: 0},
        visible: {x: 0, opacity: 1, transition: spring},
    }

    const debounce = useCallback((func: (...args: unknown[]) => void, wait: number) => {
        let timeout: NodeJS.Timeout
        return (...args: unknown[]) => {
            clearTimeout(timeout)
            timeout = setTimeout(() => func(...args), wait)
        }
    }, [])

    const getTOCWithSelect = useCallback((items: TocItem[], currentActiveAnchor: string | null): TocItem[] => {
        if (!Array.isArray(items)) {
            return []
        }
        return items.map((item) => ({
            ...item,
            isSelect:
                item.anchor === currentActiveAnchor ||
                (item.children && item.children.some((child) => child.anchor === currentActiveAnchor)),
            children: item.children ? getTOCWithSelect(item.children, currentActiveAnchor) : [],
        }))
    }, [])

    const tocWithSelect = getTOCWithSelect(toc, activeAnchor)

    const getDoms = useCallback((items: TocItem[]): HTMLElement[] => {
        const doms: HTMLElement[] = []
        const addToDoms = (items: TocItem[]) => {
            for (const item of items) {
                const dom = document.getElementById(item.anchor)
                if (dom) {
                    doms.push(dom)
                }
                if (item.children && item.children.length) {
                    addToDoms(item.children)
                }
            }
        }
        if (typeof document !== "undefined" && items.length) {
            addToDoms(items)
        }
        return doms
    }, [])

    const doms = getDoms(toc)

    useEffect(() => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        const handleScroll = debounce((scrollDom: HTMLElement) => {
            if (!scrollDom) return
            const range = window.innerHeight
            let newActiveAnchor = ""
            for (const dom of doms) {
                if (!dom) continue
                const top = dom.getBoundingClientRect().top
                if (top >= 0 && top <= range) {
                    newActiveAnchor = dom.id
                    break
                } else if (top > range) {
                    break
                } else {
                    newActiveAnchor = dom.id
                }
            }
            if (newActiveAnchor !== activeAnchor) {
                setActiveAnchor(newActiveAnchor)
            }
        }, 50)

        emitter.on("scroll", handleScroll)
        return () => {
            emitter.off("scroll", handleScroll)
        }
    }, [doms, activeAnchor, debounce])

    const setItemRef = useCallback((el: HTMLLIElement | null, isSelect: boolean | undefined) => {
        if (isSelect && el) {
            setActiveItemRef(el)
        }
    }, [])

    useEffect(() => {
        if (activeItemRef && tocRef.current) {
            const tocContainer = tocRef.current
            const containerRect = tocContainer.getBoundingClientRect()
            const activeItemRect = activeItemRef.getBoundingClientRect()

            const isItemVisible = activeItemRect.top >= containerRect.top && activeItemRect.bottom <= containerRect.bottom

            if (!isItemVisible) {
                activeItemRef.scrollIntoView({
                    behavior: "smooth",
                    block: "nearest",
                })
            }
        }
    }, [activeItemRef])

    const renderTocItems = useCallback(
        (items: TocItem[]) => {
            return items.map((item, index) => (
                <motion.div
                    key={`${item.anchor}-${index}`}
                    style={{
                        paddingLeft: `${(item.level - 2) * 12}px`,
                    }}
                    variants={itemVariants}
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-expect-error
                    ref={(el) => setItemRef(el as HTMLLIElement, item.isSelect)}
                >
                    <li
                        className={`mb-2 transition-all duration-300 ease-in-out ${
                            item.isSelect ? "font-semibold" : "font-normal"
                        }`}
                    >
                        <a
                            href={`#${item.anchor}`}
                            className={`block py-1 px-3 rounded-md transition-all duration-300 ${
                                item.isSelect
                                    ? isDark
                                        ? "text-primary bg-primary/15 shadow-[0_0_8px_rgba(var(--primary-rgb),0.3)]"
                                        : "text-primary bg-primary/10 shadow-[0_0_8px_rgba(var(--primary-rgb),0.15)]"
                                    : isDark
                                        ? "text-gray-300 hover:text-gray-100 hover:bg-gray-800/50"
                                        : "text-gray-600 hover:text-gray-900 hover:bg-gray-100/80"
                            }`}
                        >
                            {item.name}
                        </a>
                    </li>
                    <AnimatePresence>
                        {item.isSelect && item.children && item.children.length > 0 && (
                            <motion.ul initial="hidden" animate="visible" exit="hidden" variants={containerVariants}>
                                <ul className={clsx(
                                    "mt-1 ml-2 pl-3 relative transition border-l-2",
                                    isDark 
                                        ? "border-l-primary/30" 
                                        : "border-l-primary/25"
                                )}>
                                    {renderTocItems(item.children)}
                                </ul>
                            </motion.ul>
                        )}
                    </AnimatePresence>
                </motion.div>
            ))
        },
        [isDark, containerVariants, setItemRef],
    )

    const likeHandle = () => {
        likeRequest(type, targetId).then((res) => {
            if (res) {
                toast("点赞成功，感谢您的支持！")
                setLikesNum(+res)
            } else {
                toast("您已经点过赞了捏！感谢！", {type: "info"})
            }
        })
        setLiked(true)
    }

    if (isMobile) return null

    return (
        <nav className="relative">
            <ReadingProgress/>
            
            {/* 标题栏 */}
            <div className="flex items-center justify-between mb-3 mt-2 relative">
                <div className="flex items-center space-x-2">
                    <div className={clsx(
                        "w-1 h-4 rounded-full bg-gradient-to-b",
                        isDark 
                            ? "from-primary/60 to-primary/30" 
                            : "from-primary/50 to-primary/20"
                    )}/>
                    <h3 className={clsx(
                        "font-medium text-sm tracking-wide",
                        isDark ? "text-gray-200" : "text-gray-700"
                    )}>
                        目录
                    </h3>
                </div>
                <motion.div
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                >
                    <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={() => setIsCollapsed(!isCollapsed)} 
                        className={clsx(
                            "p-1.5 h-auto rounded-md transition-all duration-200",
                            "hover:bg-primary/10 hover:text-primary",
                            isDark ? "text-gray-400" : "text-gray-500"
                        )}
                    >
                        <motion.div
                            animate={{ rotate: isCollapsed ? 180 : 0 }}
                            transition={{ duration: 0.2 }}
                        >
                            <BookOpen className="w-4 h-4"/>
                        </motion.div>
                    </Button>
                </motion.div>
            </div>

            <AnimatePresence>
                {!isCollapsed && (
                    <motion.div
                        initial={{height: 0, opacity: 0}}
                        animate={{height: "auto", opacity: 1}}
                        exit={{height: 0, opacity: 0}}
                        transition={{duration: 0.3, ease: "easeInOut"}}
                        className="overflow-hidden"
                    >
                        <div
                            className={clsx(
                                "sticky h-[20em] overflow-y-auto w-56 pr-4 scroll-smooth text-sm",
                                isDark
                                    ? "scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-gray-900"
                                    : "scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100",
                            )}
                            ref={tocRef}
                        >
                            <motion.div initial="hidden" animate="visible" variants={containerVariants}>
                                <ul className="space-y-1 py-2 transition">
                                    {tocWithSelect.length > 0 && renderTocItems(tocWithSelect)}
                                </ul>
                            </motion.div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>

            {/* 操作按钮区域 */}
            <div className={clsx(
                article_font.className,
                "flex items-center justify-start space-x-4 mt-4 pt-3 relative",
                "before:absolute before:top-0 before:left-0 before:right-8 before:h-px",
                "before:bg-gradient-to-r before:from-transparent before:via-gray-300/50 before:to-transparent",
                isDark ? "before:via-gray-700/50" : "before:via-gray-300/50"
            )}>
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <motion.div
                                whileHover={{ scale: 1.05, y: -1 }}
                                whileTap={{ scale: 0.95 }}
                                className="relative"
                            >
                                <Button
                                    onClick={likeHandle}
                                    variant="ghost"
                                    size="sm"
                                    className={clsx(
                                        "flex items-center space-x-1.5 transition-all duration-200 p-1.5 h-auto rounded-lg",
                                        "hover:bg-red-50 hover:shadow-sm",
                                        liked
                                            ? "text-red-500 hover:text-red-600"
                                            : isDark
                                                ? "text-gray-400 hover:text-red-400 hover:bg-red-950/30"
                                                : "text-gray-500 hover:text-red-500",
                                    )}
                                >
                                    <Heart className={clsx("w-4 h-4 transition-all", liked && "fill-current")}/>
                                    <span className="text-sm font-medium">{likesNum}</span>
                                </Button>
                            </motion.div>
                        </TooltipTrigger>
                        <TooltipContent>
                            <p>{liked ? "你已经点过赞了，感谢支持" : "点赞"}</p>
                        </TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <motion.div
                                whileHover={{ scale: 1.05, y: -1 }}
                                whileTap={{ scale: 0.95 }}
                            >
                                <Button
                                    onClick={() => setIsCommentOpen(true)}
                                    variant="ghost"
                                    size="sm"
                                    className={clsx(
                                        "flex items-center space-x-1.5 transition-all duration-200 p-1.5 h-auto rounded-lg",
                                        "hover:bg-blue-50 hover:shadow-sm",
                                        isDark 
                                            ? "text-gray-400 hover:text-blue-400 hover:bg-blue-950/30" 
                                            : "text-gray-500 hover:text-blue-600"
                                    )}
                                >
                                    <MessageCircle className="w-4 h-4"/>
                                    <span className="text-sm font-medium">{comments}</span>
                                </Button>
                            </motion.div>
                        </TooltipTrigger>
                        <TooltipContent>
                            <p>评论</p>
                        </TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <motion.div
                                whileHover={{ scale: 1.05, y: -1 }}
                                whileTap={{ scale: 0.95 }}
                            >
                                <Button
                                    onClick={() => {
                                        navigator.clipboard.writeText(window.location.href)
                                        toast("链接已复制到剪贴板！")
                                    }}
                                    variant="ghost"
                                    size="sm"
                                    className={clsx(
                                        "transition-all duration-200 p-1.5 h-auto rounded-lg",
                                        "hover:bg-green-50 hover:shadow-sm",
                                        isDark 
                                            ? "text-gray-400 hover:text-green-400 hover:bg-green-950/30" 
                                            : "text-gray-500 hover:text-green-600"
                                    )}
                                >
                                    <Share2 className="w-4 h-4"/>
                                </Button>
                            </motion.div>
                        </TooltipTrigger>
                        <TooltipContent>
                            <p>复制链接以分享</p>
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            </div>
            <CommentModal isOpen={isCommentOpen} onClose={() => setIsCommentOpen(false)} commentId={commentId}/>
        </nav>
    )
}

