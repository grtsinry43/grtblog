"use client";

import React from 'react';
import styles from "@/styles/moment/HomePageMomentShow.module.scss";
import {ChevronDownIcon} from "@radix-ui/themes";
import {motion} from 'framer-motion';
import {clsx} from "clsx";
import Image from "next/image";
import quoteFront from "@/assets/quote-front.svg";
import {Thinking} from "@/api/thinkings";

const HomePageMomentShow = ({thinking}: { thinking: Thinking }) => {

    return (
        <div className={styles.container}>
            <div className={clsx(styles.content, "max-w-md")}>
                {/* 左上角的上引号 */}
                <motion.div
                    initial={{y: 100}}
                    animate={{y: 0}}
                    transition={{type: 'spring', stiffness: 80, damping: 20}}
                >
                    <div className={styles.quotationMark}>
                        <Image src={quoteFront} alt={"quotation mark"} width={50} height={50}/>
                    </div>
                    <span className={clsx(styles.underlineAnimation, "text-[0.8em] cursor-pointer")}>
                        {thinking.content}
                    </span>
                    <p className="mt-4 flex justify-end text-gray-400 text-[0.75rem]">
                        <span className="mr-4"> {thinking.author} </span>
                        <span className="mr-1"> 发布于 {new Date(thinking.createdAt).toLocaleDateString()} </span>
                    </p>
                </motion.div>
            </div>
            <div className={styles.arrowContainer}>
                <motion.span
                    initial={{y: 0}}
                    animate={{y: 10}}
                    transition={{type: 'spring', stiffness: 260, damping: 20}}
                >
                    <span className="text-gray-400 text-[0.75rem]">
                        下滑查看更多
                    </span>
                </motion.span>
                <motion.div
                    initial={{y: 0}}
                    animate={{y: 10}}
                    transition={{repeat: Infinity, repeatType: "reverse", duration: 1}}
                >
                    <ChevronDownIcon style={{width: 20, height: 20}}/>
                </motion.div>
            </div>
        </div>
    );
};

export default HomePageMomentShow;
