<script>
    import { fade, scale } from 'svelte/transition';
    import { cubicOut } from 'svelte/easing';

    // 定义 Props (Svelte 5 $props 语法)
    let {
        visible = true, // 控制显示/隐藏
        size = 'w-12 h-12', // 尺寸 (Tailwind 类)
        color = 'bg-ink-50 dark:bg-ink-900 border border-jade-900 dark:border-jade-100', // 颜色 (Tailwind 类)
        duration = 1200, // 翻转动画一次循环的时间 (ms)
        class: customClass = '', // 允许传入额外的 class
        text = '' // 可选的加载文本
    } = $props();
</script>

{#if visible}
    <!--
    外层容器：负责定位和 Svelte 的进出场过渡动画
    使用 transition:scale 让加载器出现时有一个缩放弹出的效果
  -->
    <div
            class="flex items-center gap-4 flex-col justify-center {customClass}"
            in:scale={{ duration: 300, easing: cubicOut, start: 0.5 }}
            out:fade={{ duration: 200 }}
    >
        <!--
      核心动画元素
      style 中使用 CSS 变量来控制动画时长
    -->
        <div
                class="{size} {color} flip-plane"
                style="--duration: {duration}ms;"
        ></div>
        {#if text}
            <span class="w-full text-sm font-serif text-ink-600 dark:text-ink-400">{text}</span>
        {/if}
    </div>
{/if}

<style>
    /*
    Tailwind 默认没有这种复杂的 3D Keyframes，
    所以我们在 style 标签中定义它。
  */
    .flip-plane {
        /* 这里的 perspective 是关键，它制造了 3D 的纵深感 */
        animation: flip-animation var(--duration) infinite ease-in-out;
    }

    @keyframes flip-animation {
        0% {
            transform: perspective(120px) rotateX(0deg) rotateY(0deg);
        }
        50% {
            /* 翻转一半时：沿 X 轴翻转 180度 */
            transform: perspective(120px) rotateX(-180.1deg) rotateY(0deg);
        }
        100% {
            /* 结束时：沿 X 轴和 Y 轴都翻转 180度 */
            transform: perspective(120px) rotateX(-180deg) rotateY(-179.9deg);
        }
    }
</style>