// // lib/plugin-system.tsx
// import React from 'react';
//
// type PluginComponent = React.ComponentType<any>;
// type PluginModule = { default: PluginComponent };
//
// const pluginRegistry = new Map<string, PluginComponent>();
// const reactProxy = new Proxy({React}, {
//     get(target, prop) {
//         return target[prop as keyof typeof target] || window.React;
//     }
// });
//
// export async function loadPlugin(url: string): Promise<void> {
//     if (pluginRegistry.has(url)) return;
//
//     try {
//         // 使用动态 import 实现模块化加载
//         const myModule = await import(/* webpackIgnore: true */ url) as PluginModule;
//
//         // 使用代理确保使用正确的 React 实例
//         const component = myModule.default;
//         const wrappedComponent = (props: any) => {
//             return React.createElement(component, {
//                 ...props,
//                 React: reactProxy.React
//             });
//         };
//
//         pluginRegistry.set(url, wrappedComponent);
//     } catch (error) {
//         console.error(`Plugin load failed: ${url}`, error);
//         throw new Error(`Failed to load plugin: ${url}`);
//     }
// }
//
// export function getPluginComponent(url: string): PluginComponent | null {
//     return pluginRegistry.get(url) || null;
// }
