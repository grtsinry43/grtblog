// // components/SafePluginLoader.tsx
// import {usePluginSystem} from '@/context/PluginContext';
// import {useEffect, useState} from 'react';
// import {PluginErrorBoundary} from "@/components/PluginErrorBoundary";
// import {loadPlugin} from "@/lib/plugin-sys";
//
// export function SafePluginLoader({pluginUrl, params}: {
//     pluginUrl: string;
//     params?: Record<string, any>;
// }) {
//     const {getPluginComponent} = usePluginSystem();
//     const [loading, setLoading] = useState(false);
//     const [error, setError] = useState<Error | null>(null);
//
//     useEffect(() => {
//         setLoading(true);
//         loadPlugin(pluginUrl)
//             .catch(err => setError(err))
//             .finally(() => setLoading(false));
//     }, [pluginUrl]);
//
//     if (error) return <div>Plugin load error: {error.message}</div>;
//     if (loading) return <div>Loading plugin...</div>;
//
//     const Component = getPluginComponent(pluginUrl);
//
//     return Component ? (
//         <PluginErrorBoundary>
//             <Component params={params}/>
//         </PluginErrorBoundary>
//     ) : null;
// }
