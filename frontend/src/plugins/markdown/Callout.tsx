import {visit} from "unist-util-visit";

// 自定义插件：解析块和行内语法
const customCalloutRemarkPlugin = () => {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    return (tree) => {
        visit(tree, (node) => {
            // 处理块级指令 :::info
            if (node.type === "containerDirective" && node.name === "info") {
                const {title} = node.attributes || {};
                node.data = {
                    hName: "InfoBlock",
                    hProperties: {
                        className: "InfoBlock",
                        title: title || "提示",
                    },
                };
                // console.log(node);
            }
            if (node.type === "containerDirective" && node.name === "warning") {
                const {title} = node.attributes || {};
                node.data = {
                    hName: "InfoBlock",
                    hProperties: {
                        className: "info-block",
                        title: title || "提示",
                    },
                };
                // console.log(node);
            }
            // console.log(tree);
        });
    };
};

export default customCalloutRemarkPlugin;
