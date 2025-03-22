import { solarizedDarkAtom, solarizedlight } from 'react-syntax-highlighter/dist/esm/styles/prism';

const customSolarizedDarkAtom = {
  ...solarizedDarkAtom,
  'code[class*="language-"]': {
    ...solarizedDarkAtom['code[class*="language-"]'],
    background: 'none', // 移除背景颜色
    fontFamily: 'inherit', // 使用继承的字体
  },
  'pre[class*="language-"]': {
    ...solarizedDarkAtom['pre[class*="language-"]'],
    background: 'none', // 移除背景颜色
    fontFamily: 'inherit', // 使用继承的字体
  },
};

const customSolarizedLightAtom = {
  ...solarizedlight,
  'code[class*="language-"]': {
    ...solarizedlight['code[class*="language-"]'],
    background: 'transparent', // 移除背景颜色
    fontFamily: 'inherit', // 使用继承的字体
  },
  'pre[class*="language-"]': {
    ...solarizedlight['pre[class*="language-"]'],
    background: 'transparent', // 移除背景颜色
    fontFamily: 'inherit', // 使用继承的字体
  },
};

// eslint-disable-next-line import/no-anonymous-default-export
export default {
  customSolarizedDarkAtom,
  customSolarizedLightAtom,
};
