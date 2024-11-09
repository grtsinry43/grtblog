export interface AddArticleApiParams {
  title?: string;
  content?: string;
  cover?: string;
  categoryId?: number;
  tags?: string;
  status?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AddArticleApiRes {
  code: number;
  msg: string;
  data: {
    id: number;
    title: string;
    content: string;
    authorId: number;
    cover: string;
    views: number;
    likes: number;
    comments: number;
    status: string;
    createdAt: string;
    updatedAt: string;
    categoryId: number;
  };
}
