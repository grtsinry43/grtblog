import React from 'react';

const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const TestDelay = async ({ id }: { id: string }) => {
  await delay(1000); // 2 seconds delay
  return <div>Comment List for {id}</div>;
};

export default TestDelay;
