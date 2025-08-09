// // components/ErrorBoundary.tsx
// import {Component, ErrorInfo, ReactNode} from 'react';
//
// interface ErrorBoundaryProps {
//     fallback?: ReactNode;
//     children: ReactNode;
// }
//
// interface ErrorBoundaryState {
//     hasError: boolean;
// }
//
// export class PluginErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
//     state = {hasError: false};
//
//     static getDerivedStateFromError() {
//         return {hasError: true};
//     }
//
//     componentDidCatch(error: Error, info: ErrorInfo) {
//         console.error('Plugin Error:', error, info);
//     }
//
//     render() {
//         return this.state.hasError
//             ? this.props.fallback || <div> 诶呀，出错了。</div>
//             : this.props.children;
//     }
// }
