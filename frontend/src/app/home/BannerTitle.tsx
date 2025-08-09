"use client"

import {useState, useEffect} from "react"
import {clsx} from "clsx"
import {noto_sans_sc, playwrite_us_modern, varela_round} from '@/app/fonts/font'
import {useWebsiteInfo} from "@/app/website-info-provider"
import {motion} from "framer-motion"

// 字符逐个淡入动画组件
function AnimatedText({ text, className, delay = 0 }: { text: string; className: string; delay?: number }) {
    const characters = text.split('')
    
    return (
        <span className={className}>
            {characters.map((char, index) => (
                <motion.span
                    key={index}
                    initial={{ opacity: 0, y: 20, filter: "blur(4px)" }}
                    animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                    transition={{
                        duration: 0.5,
                        delay: delay + index * 0.05,
                        ease: [0.25, 0.46, 0.45, 0.94]
                    }}
                    className="inline-block"
                >
                    {char === ' ' ? '\u00A0' : char}
                </motion.span>
            ))}
        </span>
    )
}

// 单词级别的淡入动画组件
function AnimatedWords({ text, className, delay = 0 }: { text: string; className: string; delay?: number }) {
    const words = text.split(' ')
    
    return (
        <span className={className}>
            {words.map((word, index) => (
                <motion.span
                    key={index}
                    initial={{ opacity: 0, scale: 0.8, y: 15 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    transition={{
                        duration: 0.6,
                        delay: delay + index * 0.15,
                        ease: "easeOut"
                    }}
                    className="inline-block mr-2"
                >
                    {word}
                </motion.span>
            ))}
        </span>
    )
}

export default function BannerTitle() {
    const websiteInfo = useWebsiteInfo()
    const [visible, setVisible] = useState(false)

    useEffect(() => {
        setVisible(true)
    }, [])

    return (
        <motion.div
            initial={{opacity: 0}}
            animate={{opacity: visible ? 1 : 0}}
            transition={{duration: 0.8}}
            className="flex flex-col justify-center flex-1 p-5 sm:p-10 space-y-5 relative"
        >
            {/* 装饰性背景元素 */}
            <motion.div 
                className="absolute top-0 left-0 w-2 h-16 bg-gradient-to-b from-primary/70 to-primary/10 rounded-full"
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 64, opacity: 1 }}
                transition={{ delay: 0.5, duration: 0.8 }}
            />
            <motion.div 
                className="absolute top-6 left-6 w-1 h-10 bg-gradient-to-b from-accent/50 to-transparent rounded-full"
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 40, opacity: 1 }}
                transition={{ delay: 0.8, duration: 0.6 }}
            />
            
            {/* 主标题 */}
            <motion.div
                className="relative"
                initial={{y: 20}}
                animate={{y: 0}}
                transition={{delay: 0.2, duration: 0.5}}
            >
                <div className={clsx(
                    varela_round.className,
                    "text-3xl md:text-4xl lg:text-5xl font-bold",
                    "text-foreground",
                    "relative tracking-tight"
                )}>
                    <AnimatedText 
                        text={websiteInfo.HOME_TITLE}
                        className=""
                        delay={0.3}
                    />
                </div>
                {/* 标题下方装饰线 */}
                <motion.div 
                    className="h-0.5 bg-gradient-to-r from-primary via-primary/50 to-transparent mt-3 rounded-full"
                    initial={{width: 0, opacity: 0}}
                    animate={{width: "50%", opacity: 1}}
                    transition={{delay: 1.2, duration: 0.8, ease: "easeOut"}}
                />
            </motion.div>

            {/* 英文标语 */}
            <motion.div
                className="relative pl-6"
                initial={{y: 20, opacity: 0}}
                animate={{y: 0, opacity: 1}}
                transition={{delay: 0.5, duration: 0.5}}
            >
                {/* 小装饰点 */}
                <motion.div 
                    className="absolute -left-1 top-1/2 transform -translate-y-1/2 w-1.5 h-1.5 bg-primary/60 rounded-full"
                    initial={{ scale: 0, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ delay: 1.5, duration: 0.3 }}
                />
                
                <div className={clsx(
                    playwrite_us_modern.className,
                    "text-lg md:text-xl lg:text-2xl",
                    "text-primary italic",
                    "font-normal tracking-wide",
                    "leading-relaxed"
                )}>
                    <AnimatedWords
                        text={websiteInfo.HOME_SLOGAN_EN}
                        className=""
                        delay={1.5}
                    />
                </div>
            </motion.div>

            {/* 中文标语 */}
            <motion.div
                className="relative pl-10"
                initial={{y: 20, opacity: 0}}
                animate={{y: 0, opacity: 1}}
                transition={{delay: 0.8, duration: 0.5}}
            >
                <div className="flex items-center space-x-3">
                    {/* 装饰性图标 */}
                    <motion.div 
                        className="w-1.5 h-1.5 bg-accent rounded-full ring-2 ring-accent/20"
                        initial={{scale: 0, rotate: -180}}
                        animate={{scale: 1, rotate: 0}}
                        transition={{delay: 2.5, duration: 0.4, ease: "backOut"}}
                    />
                    <div className={clsx(
                        noto_sans_sc.className,
                        "text-base md:text-lg lg:text-xl",
                        "text-muted-foreground font-light",
                        "tracking-wider leading-relaxed"
                    )}>
                        <AnimatedText
                            text={websiteInfo.HOME_SLOGAN}
                            className=""
                            delay={2.8}
                        />
                    </div>
                </div>
            </motion.div>

            {/* 底部装饰性元素 */}
            <motion.div 
                className="flex items-center space-x-2 pl-6 pt-6"
                initial={{opacity: 0, x: -20}}
                animate={{opacity: 1, x: 0}}
                transition={{delay: 4, duration: 0.8}}
            >
                <motion.div 
                    className="w-8 h-px bg-gradient-to-r from-primary/60 to-primary/20 rounded-full"
                    initial={{ scaleX: 0 }}
                    animate={{ scaleX: 1 }}
                    transition={{ delay: 4.2, duration: 0.6 }}
                />
                <motion.div 
                    className="w-1 h-1 bg-primary/40 rounded-full"
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: 4.4, duration: 0.3 }}
                />
                <motion.div 
                    className="w-4 h-px bg-gradient-to-r from-accent/40 to-transparent rounded-full"
                    initial={{ scaleX: 0 }}
                    animate={{ scaleX: 1 }}
                    transition={{ delay: 4.6, duration: 0.4 }}
                />
                <motion.div 
                    className="w-px h-px bg-muted-foreground/30 rounded-full"
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ delay: 4.8, duration: 0.2 }}
                />
            </motion.div>
        </motion.div>
    )
}