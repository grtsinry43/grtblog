"use client"

import {useState, useEffect} from "react"
import {clsx} from "clsx"
import {noto_sans_sc, playwrite_us_modern, varela_round} from '@/app/fonts/font'
import {useWebsiteInfo} from "@/app/website-info-provider"
import {ReactTyped} from "react-typed"
import {motion} from "framer-motion"

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
            className="flex flex-col justify-center flex-1 p-10 space-y-4"
        >
            <motion.div
                className={clsx(
                    varela_round.className,
                    "text-3xl md:text-4xl lg:text-5xl font-bold"
                )}
                initial={{y: 20}}
                animate={{y: 0}}
                transition={{delay: 0.2, duration: 0.5}}
            >
                <ReactTyped
                    strings={[websiteInfo.HOME_TITLE]}
                    typeSpeed={50}
                    startDelay={300}
                    cursorChar="_"
                    fadeOut={true}
                    onComplete={(self) => {
                        // eslint-disable-next-line @typescript-eslint/no-unused-expressions
                        self.cursor && self.cursor.remove();
                    }}
                />
            </motion.div>

            <motion.div
                className={clsx(
                    playwrite_us_modern.className,
                    "text-xl md:text-2xl lg:text-3xl text-primary"
                )}
                initial={{y: 20, opacity: 0}}
                animate={{y: 0, opacity: 1}}
                transition={{delay: 0.5, duration: 0.5}}
            >
                <ReactTyped
                    strings={[websiteInfo.HOME_SLOGAN_EN]}
                    typeSpeed={30}
                    startDelay={1500}
                    cursorChar="|"
                    fadeOut={true}
                    onComplete={(self) => {
                        // eslint-disable-next-line @typescript-eslint/no-unused-expressions
                        self.cursor && self.cursor.remove();
                    }}
                />
            </motion.div>

            <motion.div
                className={clsx(
                    noto_sans_sc.className,
                    "text-xl md:text-2xl mt-3 text-muted-foreground"
                )}
                initial={{y: 20, opacity: 0}}
                animate={{y: 0, opacity: 1}}
                transition={{delay: 0.8, duration: 0.5}}
            >
                <ReactTyped
                    strings={[websiteInfo.HOME_SLOGAN]}
                    typeSpeed={40}
                    startDelay={2500}
                    cursorChar="❯"
                    fadeOut={true}
                    onComplete={(self) => {
                        // eslint-disable-next-line @typescript-eslint/no-unused-expressions
                        self.cursor && self.cursor.remove();
                    }}
                />
            </motion.div>
        </motion.div>
    )
}