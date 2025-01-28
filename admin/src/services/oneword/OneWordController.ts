import { OneWordResponse } from '@/services/oneword/typings';

const refreshOneWord = async (): Promise<OneWordResponse> => {
  const res = await fetch('https://v1.hitokoto.cn');
  return await res.json();
};

export default {
  refreshOneWord,
};
