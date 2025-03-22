"use client"
import Link from "next/link"
import {motion} from "framer-motion"

const TagItemCard = ({name, count, index}: { name: string; count: number; index: number }) => {
    // 精美的颜色组合
    const colors = [
        {
            bg: "bg-blue-50 dark:bg-blue-900/30",
            text: "text-blue-600 dark:text-blue-300",
            hover: "hover:bg-blue-100 dark:hover:bg-blue-800/40",
            border: "border-blue-200 dark:border-blue-800/50",
        },
        {
            bg: "bg-purple-50 dark:bg-purple-900/30",
            text: "text-purple-600 dark:text-purple-300",
            hover: "hover:bg-purple-100 dark:hover:bg-purple-800/40",
            border: "border-purple-200 dark:border-purple-800/50",
        },
        {
            bg: "bg-pink-50 dark:bg-pink-900/30",
            text: "text-pink-600 dark:text-pink-300",
            hover: "hover:bg-pink-100 dark:hover:bg-pink-800/40",
            border: "border-pink-200 dark:border-pink-800/50",
        },
        {
            bg: "bg-emerald-50 dark:bg-emerald-900/30",
            text: "text-emerald-600 dark:text-emerald-300",
            hover: "hover:bg-emerald-100 dark:hover:bg-emerald-800/40",
            border: "border-emerald-200 dark:border-emerald-800/50",
        },
        {
            bg: "bg-amber-50 dark:bg-amber-900/30",
            text: "text-amber-600 dark:text-amber-300",
            hover: "hover:bg-amber-100 dark:hover:bg-amber-800/40",
            border: "border-amber-200 dark:border-amber-800/50",
        },
        {
            bg: "bg-cyan-50 dark:bg-cyan-900/30",
            text: "text-cyan-600 dark:text-cyan-300",
            hover: "hover:bg-cyan-100 dark:hover:bg-cyan-800/40",
            border: "border-cyan-200 dark:border-cyan-800/50",
        },
        {
            bg: "bg-indigo-50 dark:bg-indigo-900/30",
            text: "text-indigo-600 dark:text-indigo-300",
            hover: "hover:bg-indigo-100 dark:hover:bg-indigo-800/40",
            border: "border-indigo-200 dark:border-indigo-800/50",
        },
    ]

    const color = colors[index % colors.length]

    return (
        <motion.div
            initial={{opacity: 0, y: 10}}
            animate={{opacity: 1, y: 0}}
            transition={{
                type: "spring",
                stiffness: 100,
                mass: 0.5,
                delay: index * 0.02,
            }}
            whileHover={{
                scale: 1.05,
                y: -2,
                transition: {duration: 0.2},
            }}
        >
            <Link
                href={`/tags/${name}`}
                className={`group backdrop-blur-lg flex items-center m-2 px-3 py-1.5 rounded-sm ${color.bg} ${color.hover} ${color.border} border transition-all duration-200 shadow-sm hover:shadow`}
            >
                <span className={`text-xs font-medium ${color.text}`}>{name}</span>

                <span
                    className="ml-1.5 px-1.5 py-0.5 text-[10px] rounded-full bg-white/50 dark:bg-slate-800/50 text-slate-500 dark:text-slate-400 group-hover:bg-white dark:group-hover:bg-slate-700 transition-colors">
          {count}
        </span>
            </Link>
        </motion.div>
    )
}

export default TagItemCard

