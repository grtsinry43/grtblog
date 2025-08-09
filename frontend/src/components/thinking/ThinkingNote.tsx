"use client"

import {motion} from "framer-motion"
import type {Thinking} from "@/api/thinkings"
import {Calendar, User} from "lucide-react"

interface ThinkingNoteProps {
    thinking: Thinking
    index: number
}

const colors = [
    "bg-amber-50 border-amber-200 dark:bg-amber-950/40 dark:border-amber-800/70",
    "bg-emerald-50 border-emerald-200 dark:bg-emerald-950/40 dark:border-emerald-800/70",
    "bg-sky-50 border-sky-200 dark:bg-sky-950/40 dark:border-sky-800/70",
    "bg-rose-50 border-rose-200 dark:bg-rose-950/40 dark:border-rose-800/70",
    "bg-violet-50 border-violet-200 dark:bg-violet-950/40 dark:border-violet-800/70",
]

const textColors = [
    "text-amber-800 dark:text-amber-200",
    "text-emerald-800 dark:text-emerald-200",
    "text-sky-800 dark:text-sky-200",
    "text-rose-800 dark:text-rose-200",
    "text-violet-800 dark:text-violet-200",
]

export default function ThinkingNote({thinking, index}: ThinkingNoteProps) {
    const colorIndex = index % colors.length
    // Create a more subtle rotation between -2 and 2 degrees
    const rotation = (Math.random() * 4 - 2).toFixed(1)

    return (
        <motion.div
            initial={{opacity: 0, y: 50}}
            animate={{opacity: 1, y: 0}}
            transition={{
                duration: 0.6,
                delay: index * 0.08,
                ease: "easeOut",
            }}
            whileHover={{
                scale: 1.03,
                rotate: 0,
                boxShadow: "0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)",
            }}
            style={{rotate: `${rotation}deg`}}
            className="will-change-transform"
        >
            <div
                className={`${colors[colorIndex]} p-6 rounded-xl border backdrop-blur-sm 
        shadow-md dark:shadow-lg break-inside-avoid mb-6 relative overflow-hidden`}
            >
                {/* Decorative elements */}
                <div
                    className="absolute -top-6 -right-6 w-12 h-12 rounded-full bg-gradient-to-br from-white/20 to-transparent dark:from-white/5"></div>
                <div
                    className="absolute -bottom-8 -left-8 w-16 h-16 rounded-full bg-gradient-to-tr from-black/5 to-transparent dark:from-white/5"></div>

                {/* Quote marks */}
                <div className="absolute top-2 left-2 text-4xl opacity-10 font-serif">&#34;</div>

                {/* Content */}
                <div className="relative">
                    <p className={`${textColors[colorIndex]} text-base leading-relaxed mb-6 font-medium relative z-10`}>
                        {thinking.content}
                    </p>

                    <div
                        className="flex justify-between items-center text-sm text-gray-600 dark:text-gray-300 pt-3 border-t border-gray-200 dark:border-gray-700/30">
                        <div className="flex items-center gap-1.5">
                            <User size={14}/>
                            <span>{thinking.author}</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                            <Calendar size={14}/>
                            <span>{new Date(thinking.createdAt).toLocaleDateString()}</span>
                        </div>
                    </div>
                </div>
            </div>
        </motion.div>
    )
}

