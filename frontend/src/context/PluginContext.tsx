// // context/PluginContext.tsx
// import React, {createContext, useContext, useEffect, useState} from 'react';
// import {loadPlugin, getPluginComponent} from '@/lib/plugin-sys';
// import {getPluginList, PluginFetchItem} from "@/api/plugins";
//
// const PluginContext = createContext({
//     plugins: new Map<string, string>(),
//     loadPlugin: async (url: string) => {
//     },
//     getPluginComponent: (name: string): any => {
//     }
// });
//
// export function PluginProvider({children}: { children: React.ReactNode }) {
//     const [plugins, setPlugins] = useState(new Map());
//
//     useEffect(() => {
//         getPluginList().then(plugins => {
//             plugins.forEach((p: PluginFetchItem) => {
//                 loadPlugin(p.endpoint).then(() => {
//                     setPlugins(prev => new Map(prev).set(p.name, p.endpoint));
//                 });
//             });
//         });
//     }, []);
//
//     const value = {
//         plugins,
//         loadPlugin: async (url: string) => {
//             await loadPlugin(url);
//             setPlugins(prev => new Map(prev).set(url, url));
//         },
//         getPluginComponent: (name: string) => {
//             console.log("获取插件组件", name);
//             console.log("插件列表", plugins);
//             const url = Array.from(plugins.values()).find(u => u.includes(name));
//             return url ? getPluginComponent(url) : null;
//         }
//     };
//
//     return (
//         <PluginContext.Provider value={value}>
//             {children}
//         </PluginContext.Provider>
//     );
// }
//
// export const usePluginSystem = () => useContext(PluginContext);
