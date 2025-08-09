// "use client";
//
// import React, {Suspense} from 'react';
// import {PluginProvider, usePluginSystem} from "@/context/PluginContext";
// import {PluginErrorBoundary} from "@/components/PluginErrorBoundary";
// import {SafePluginLoader} from "@/components/PluginLoader";
//
// const PluginClient = () => {
//     // const {getPluginComponent} = usePluginSystem();
//     // const pluginName = 'test-plugin';
//     // const PluginComponent = getPluginComponent(pluginName);
//     return (
//         <div>
//             <PluginProvider>
//                 <PluginErrorBoundary>
//                     <Suspense fallback={<div>Loading...</div>}>
//                         <SafePluginLoader
//                             pluginUrl={"http://localhost:8080/plugins/netease/playlist/js"}
//                             params={{}}
//                         />
//                     </Suspense>
//                 </PluginErrorBoundary>
//             </PluginProvider>
//         </div>
//     );
// };
//
// export default PluginClient;
