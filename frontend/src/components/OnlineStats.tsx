"use client";

import React, {useEffect, useState} from "react";
import {io, Socket} from "socket.io-client";
import {useAppDispatch} from "@/redux/hooks";
import {usePathname} from "next/navigation";
import {getPageView} from "@/api/pageView";
import channel from "@/utils/channel";
import emitter from "@/utils/eventBus";

const url = process.env.NEXT_PUBLIC_SOCKET_IO_URL;

const OnlineStats = () => {
    const [socket, setSocket] = useState<Socket | null>(null); // Socket.IO 实例
    const [pageViewCount, setPageViewCount] = useState(0); // 当前页面在线人数
    const [totalOnlineCount, setTotalOnlineCount] = useState(0); // 总在线人数
    const param = usePathname();

    const dispatch = useAppDispatch();

    // 初始化 Socket.IO 连接
    useEffect(() => {

        const newSocket = io(url);
        setSocket(newSocket);

        getPageView().then((res) => {

            dispatch({
                type: "onlineCount/initPageView",
                payload: res
            })
        });

        // 监听总在线人数事件
        newSocket.on("totalOnlineCount", (count: number) => {
            setTotalOnlineCount(count);

            dispatch({type: "onlineCount/updateOnlineCount", payload: count});
        });

        // 监听页面在线人数事件
        newSocket.on("pageViewCount", (page, count) => {
            if (page === param) {
                setPageViewCount(count);

            }

            dispatch({
                type: "onlineCount/updatePageView", payload: {
                    name: page,
                    count: count
                }
            });
        });

        // 用于发送更新通知
        newSocket.on("updateNotification", (content) => {
            console.log(content);
            channel.port2.postMessage({
                content: content,
                publishAt: new Date().toISOString(),
            });
            console.log("send message");
        });

        // 用于更新站长实时在线动态
        newSocket.on("authorStatus", (content) => {
            console.log(content);
            emitter.emit("authorStatus", content);
        })

        // 在组件卸载时关闭连接
        return () => {
            newSocket.disconnect();
        };
    }, [param, dispatch, totalOnlineCount, pageViewCount]);

    // 发送当前页面信息
    useEffect(() => {
        if (socket) {
            socket.emit("enterPage", param);
        }
    }, [socket, param]);

    return (
        <></>
        // <div style={{padding: "20px", fontFamily: "Arial, sans-serif"}}>
        //     <h1> 实时在线统计 </h1>
        //     <div>
        //         <p> 当前页面在线人数：{pageViewCount}</p>
        //         <p> 全站在线人数：{totalOnlineCount}</p>
        //     </div>
        // </div>
    );
};

export default OnlineStats;
