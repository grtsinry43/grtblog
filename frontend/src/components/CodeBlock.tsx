"use client"

import React, {useEffect, useState} from 'react'
import {CopyToClipboard} from 'react-copy-to-clipboard'
import {Prism as SyntaxHighlighter} from 'react-syntax-highlighter'
import {useTheme} from 'next-themes'
import {cn} from '@/lib/utils'
import {AnimatePresence, motion} from 'framer-motion'
import {Check, ChevronDown, ChevronUp, Copy, Terminal} from 'lucide-react'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {jetbrains_mono} from '@/app/fonts/font'
import fallbackTheme from "@/components/code/fallbackTheme";
import customTheme from "@/components/code/customTheme";
import {ScrollArea} from "@radix-ui/themes";

interface CodeBlockProps {
    language: string
    value: string
    showLineNumbers?: boolean
    className?: string
    initialVisibleLines?: number
}

const CodeBlock: React.FC<CodeBlockProps> = ({
                                                 language,
                                                 value,
                                                 showLineNumbers = true,
                                                 initialVisibleLines = 10,
                                             }) => {
    const [copied, setCopied] = useState(false)
    const {resolvedTheme} = useTheme()
    const isDark = resolvedTheme === 'dark'
    const [expanded, setExpanded] = useState(false)
    const [shouldShowExpand, setShouldShowExpand] = useState(false)
    const [bgClassName, setBgClassName] = useState('')

    const codeLines = value.split('\n')

    useEffect(() => {
        setShouldShowExpand(codeLines.length > initialVisibleLines)
        setBgClassName(isDark ? "bg-zinc-900 border-zinc-800" : "bg-zinc-50 border-zinc-200")

    }, [codeLines.length, initialVisibleLines, isDark])

    const displayedCode = expanded
        ? value
        : codeLines.slice(0, initialVisibleLines).join('\n')

    const handleCopy = () => {
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    const toggleExpand = () => {
        setExpanded(!expanded)
    }

    const currentTheme = isDark ? customTheme.customSolarizedDarkAtom : customTheme.customSolarizedLightAtom || fallbackTheme

    return (
        <div className={cn(
            "relative rounded-lg overflow-hidden border my-4 mx-2",
            bgClassName,
            jetbrains_mono.className
        )}>
            <div className="flex items-center justify-between px-2 border-b border-border">
                <div className="flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-muted-foreground"/>
                    <Badge variant="outline" className="text-xs font-medium uppercase">
                        {language}
                    </Badge>
                </div>
                <CopyToClipboard text={value} onCopy={handleCopy}>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 rounded-full"
                    >
                        <AnimatePresence mode="wait" initial={false}>
                            {copied ? (
                                <motion.div
                                    key="check"
                                    initial={{scale: 0.8, opacity: 0}}
                                    animate={{scale: 1, opacity: 1}}
                                    exit={{scale: 0.8, opacity: 0}}
                                    transition={{duration: 0.15}}
                                >
                                    <Check className="h-4 w-4 text-green-500"/>
                                </motion.div>
                            ) : (
                                <motion.div
                                    key="copy"
                                    initial={{scale: 0.8, opacity: 0}}
                                    animate={{scale: 1, opacity: 1}}
                                    exit={{scale: 0.8, opacity: 0}}
                                    transition={{duration: 0.15}}
                                >
                                    <Copy className="h-4 w-4 text-muted-foreground"/>
                                </motion.div>
                            )}
                        </AnimatePresence>
                    </Button>
                </CopyToClipboard>
            </div>

            <ScrollArea className="w-full" scrollbars="horizontal">
                <div className={cn("relative", jetbrains_mono.className)}>
                    <SyntaxHighlighter
                        language={language}
                        style={currentTheme}
                        showLineNumbers={showLineNumbers}
                        wrapLines={false}
                        customStyle={{
                            margin: 0,
                            padding: '0.7rem',
                            fontSize: '12px',
                        }}
                        lineNumberStyle={{
                            minWidth: '2.5em',
                            paddingRight: '1em',
                            marginRight: '1em',
                            textAlign: 'right',
                            userSelect: 'none',
                            opacity: 0.5,
                            borderRight: isDark ? '1px solid #333' : '1px solid #eaeaea',
                            position: 'sticky',
                            left: 0,
                        }}
                    >
                        {displayedCode}
                    </SyntaxHighlighter>
                </div>
            </ScrollArea>

            {shouldShowExpand && (
                <div className="flex justify-center border-t border-border">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={toggleExpand}
                        className="w-full rounded-none flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
                    >
                        {expanded ? (
                            <>
                                <ChevronUp className="h-3 w-3"/>
                                <span>收起更多</span>
                            </>
                        ) : (
                            <>
                                <ChevronDown className="h-3 w-3"/>
                                <span>展示全部 {codeLines.length} 行代码内容</span>
                            </>
                        )}
                    </Button>
                </div>
            )}
        </div>
    )
}

export default CodeBlock
