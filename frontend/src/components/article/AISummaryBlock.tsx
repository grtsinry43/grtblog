"use client"

import type React from "react"
import {useState, useRef, useEffect} from "react"
import ReactMarkdown from "react-markdown"
import {Sparkles, ChevronDown, ChevronUp, Pause, Play} from "lucide-react"
import styles from "@/styles/PostPage.module.scss"
import rehypeCallouts from "rehype-callouts"
import remarkGfm from "remark-gfm"
import remarkBreaks from "remark-breaks"
import InlineCodeBlock from "@/components/InlineCodeBlock"
import CodeBlock from "@/components/CodeBlock"
import {clsx} from "clsx"
import TableView from "@/components/article/TableView"
import Link from "next/link"
import rehypeRaw from "rehype-raw"

interface AiSummaryBlockProps {
    aiSummary: string
}

const AiSummaryBlock: React.FC<AiSummaryBlockProps> = ({aiSummary}) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const contentRef = useRef<HTMLDivElement>(null)
    const [showGradient, setShowGradient] = useState(true)

    // Animation states
    const [displayedContent, setDisplayedContent] = useState("")
    const [isTyping, setIsTyping] = useState(true)
    const [typingComplete, setTypingComplete] = useState(false)
    const typingSpeed = 5

    // Cursor blinking effect
    const [showCursor, setShowCursor] = useState(true)

    useEffect(() => {
        if (contentRef.current) {
            setShowGradient(contentRef.current.scrollHeight > 100)
        }
    }, [contentRef.current, displayedContent])

    // Typing animation effect
    useEffect(() => {
        if (!aiSummary || typingComplete) return

        const currentIndex = displayedContent.length

        if (currentIndex >= aiSummary.length) {
            setTypingComplete(true)
            return
        }

        if (!isTyping) return

        const timer = setTimeout(() => {
            setDisplayedContent(aiSummary.substring(0, currentIndex + 1))
        }, typingSpeed)

        return () => clearTimeout(timer)
    }, [aiSummary, displayedContent, isTyping, typingComplete, typingSpeed])

    // Blinking cursor effect
    useEffect(() => {
        if (typingComplete) return

        const cursorInterval = setInterval(() => {
            setShowCursor((prev) => !prev)
        }, 500)

        return () => clearInterval(cursorInterval)
    }, [typingComplete])

    const handleToggleTyping = () => {
        if (typingComplete) {
            // Reset animation if already complete
            setDisplayedContent("")
            setTypingComplete(false)
            setIsTyping(true)
        } else {
            // Pause/resume typing
            setIsTyping(!isTyping)
        }
    }

    const handleSkipAnimation = () => {
        setDisplayedContent(aiSummary)
        setTypingComplete(true)
        setIsTyping(false)
    }

    if (!aiSummary) {
        return null
    }

    return (
        <div
            className="bg-gradient-to-r from-purple-50/80 to-indigo-50/80 dark:from-purple-950/30 dark:to-indigo-950/30 rounded-md p-3 mb-8 shadow-sm"
            style={{
                border: "1px solid rgba(var(--foreground), 0.06)",
                backdropFilter: "blur(8px)",
            }}
        >
            <div className="flex items-center mb-3">
                <Sparkles className="w-5 h-5 text-purple-500/70 dark:text-purple-400/70 mr-2"/>
                <h2 className="text-sm font-medium text-gray-700 dark:text-gray-300 flex-grow">AI Summary</h2>
                <div className="flex items-center gap-2 mr-2">
                    {!typingComplete && (
                        <>
                            <button
                                onClick={handleToggleTyping}
                                className="text-purple-500/70 dark:text-purple-400/70 hover:text-purple-600 dark:hover:text-purple-300 transition-colors duration-200"
                                title={isTyping ? "Pause typing" : "Resume typing"}
                            >
                                {isTyping ? <Pause className="w-3.5 h-3.5"/> : <Play className="w-3.5 h-3.5"/>}
                            </button>
                            <button
                                onClick={handleSkipAnimation}
                                className="text-purple-500/70 dark:text-purple-400/70 hover:text-purple-600 dark:hover:text-purple-300 transition-colors duration-200 text-xs"
                                title="Skip animation"
                            >
                                Skip
                            </button>
                        </>
                    )}
                </div>
                <div className="opacity-50 text-[0.65em] self-end sm:self-center text-gray-600 dark:text-gray-400">
                    Powered By DeepSeek-R1
                </div>
            </div>
            <div className="relative text-xs">
                <div
                    ref={contentRef}
                    className={clsx(
                        "prose prose-sm dark:prose-invert max-w-none overflow-hidden transition-all duration-300 ease-in-out text-gray-700 dark:text-gray-300",
                        isExpanded ? "max-h-[1000px]" : "max-h-[100px]",
                        !isExpanded && showGradient && "mask-image-fade",
                    )}
                >
                    <ReactMarkdown
                        className={clsx(styles.markdown, "text-xs leading-relaxed")}
                        rehypePlugins={[rehypeCallouts, rehypeRaw]}
                        remarkPlugins={[remarkGfm, remarkBreaks]}
                        components={{
                            code({inline, className, children, ...props}) {
                                const match = /language-(\w+)/.exec(className || "")
                                if (!match) {
                                    return <InlineCodeBlock {...props}>{children}</InlineCodeBlock>
                                }
                                return inline ? (
                                    <InlineCodeBlock {...props}>{children}</InlineCodeBlock>
                                ) : (
                                    <CodeBlock language={match[1]} value={String(children).replace(/\n$/, "")}/>
                                )
                            },
                            a({...props}) {
                                return (
                                    <Link
                                        style={{color: "#8a9eff"}}
                                        className={clsx(styles.underlineAnimation, styles.glowAnimation, "ml-0.5 mr-0.5")}
                                        {...props}
                                    />
                                )
                            },
                            p({...props}) {
                                return <p
                                    className={clsx(styles.paragraph, "text-xs leading-relaxed my-1.5")} {...props} />
                            },
                            table({...props}) {
                                return <TableView {...props} />
                            },
                            strong({...props}) {
                                return <strong className={clsx(styles.bold, "font-medium")} {...props} />
                            },
                            em({...props}) {
                                return <em className={styles.italic} {...props} />
                            },
                            blockquote({...props}) {
                                return <blockquote
                                    className={clsx(styles.blockquote, "text-xs border-l-2 pl-3 my-2")} {...props} />
                            },
                            h1({...props}) {
                                return <h1
                                    className={clsx(styles.heading1, "text-sm font-medium mt-3 mb-1.5")} {...props} />
                            },
                            h2({...props}) {
                                return <h2
                                    className={clsx(styles.heading2, "text-xs font-medium mt-2.5 mb-1")} {...props} />
                            },
                            h3({...props}) {
                                return <h3
                                    className={clsx(styles.heading3, "text-xs font-medium mt-2 mb-1")} {...props} />
                            },
                            h4({...props}) {
                                return <h4
                                    className={clsx(styles.heading4, "text-xs font-medium mt-2 mb-1")} {...props} />
                            },
                            ul({...props}) {
                                return <ul className="text-xs my-1.5 pl-4" {...props} />
                            },
                            ol({...props}) {
                                return <ol className="text-xs my-1.5 pl-4" {...props} />
                            },
                            li({...props}) {
                                return <li className="text-xs my-0.5" {...props} />
                            },
                        }}
                    >
                        {displayedContent}
                    </ReactMarkdown>
                    {!typingComplete && showCursor && (
                        <span className="typing-cursor inline-block ml-0.5 h-3 align-middle">|</span>
                    )}
                </div>
            </div>
            {showGradient && (
                <button
                    className="mt-2 flex items-center text-purple-500/70 dark:text-purple-400/70 hover:text-purple-600 dark:hover:text-purple-300 transition-colors duration-200 text-xs"
                    onClick={() => setIsExpanded(!isExpanded)}
                >
                    {isExpanded ? (
                        <>
                            <ChevronUp className="w-3.5 h-3.5 mr-1"/>
                            收起
                        </>
                    ) : (
                        <>
                            <ChevronDown className="w-3.5 h-3.5 mr-1"/>
                            展开
                        </>
                    )}
                </button>
            )}
        </div>
    )
}

export default AiSummaryBlock

