"use client"

import {motion} from "framer-motion"
import type {Thinking} from "@/api/thinkings"

interface ThinkingNoteProps {
    thinking: Thinking
    index: number
}

const colors = [
    "bg-yellow-200 dark:bg-yellow-700 dark:opacity-70",
    "bg-green-200 dark:bg-green-700 dark:opacity-70",
    "bg-blue-200 dark:bg-blue-700 dark:opacity-70",
    "bg-pink-200 dark:bg-pink-700 dark:opacity-70",
    "bg-purple-200 dark:bg-purple-700 dark:opacity-70",
]

export default function ThinkingNote({thinking, index}: ThinkingNoteProps) {
    const colorIndex = index % colors.length
    const rotation = Math.random() * 6 - 3

    return (
        <motion.div
            initial={{opacity: 0, y: 50}}
            animate={{opacity: 1, y: 0}}
            transition={{duration: 0.5, delay: index * 0.1}}
            style={{rotate: `${rotation}deg`}}
        >
            <div
                className={`${colors[colorIndex]} p-6 rounded-lg shadow-md transform hover:scale-105 transition-transform duration-200 break-inside-avoid mb-6`}>
                <p className="text-gray-800 dark:text-gray-50 mb-4">{thinking.content}</p>
                <div className="flex justify-between items-center text-sm text-gray-600 dark:text-gray-200">
                    <span>{thinking.author}</span>
                    <span>{new Date(thinking.createdAt).toLocaleDateString()}</span>
                </div>
            </div>
        </motion.div>
    )
}

