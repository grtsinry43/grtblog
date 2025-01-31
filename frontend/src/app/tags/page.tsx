import React from 'react';
import TagCloudBackground from '@/components/tags/TagCloudBackground';
import {Container} from '@radix-ui/themes';
import {getAllTags} from '@/api/tag';
import {noto_sans_sc, varela_round} from '@/app/fonts/font';
import TagItemCard from '@/components/tags/TagItemCard';
import {Metadata} from "next";

type Tag = {
    tagId: number;
    tagName: string;
    articleCount: number;
}

export const metadata: Metadata = {
    title: '标签',
    description: '这是所有标签的列表',
}

const Page = async () => {
    const tags: Tag[] = await getAllTags({next: {revalidate: 60}});
    return (
        <div>
            <Container size={'4'} style={{
                padding: '2em 20px',
                scrollBehavior: 'smooth',
                zIndex: 1,
                position: 'relative',
            }}>
                <div className="title relative">
                    <h1 style={{
                        fontSize: '2em',
                        fontWeight: 'bolder',
                        width: 'fit-content',
                        position: 'relative',
                        margin: '1em auto',
                    }} className={noto_sans_sc.className}>
                        标签
                    </h1>
                    <span
                        className={varela_round.className}
                        style={{
                            fontSize: '3em',
                            fontWeight: 'bolder',
                            color: 'rgba(var(--foreground), 0.1)',
                            position: 'absolute',
                            top: '0',
                            left: '50%',
                            transform: 'translateX(-50%)',
                        }}>TAGS</span>
                </div>
                {/*<div className="search">*/}
                {/*    <TextField.Root*/}
                {/*        placeholder="搜索标签"*/}
                {/*        className={noto_sans_sc.className}*/}
                {/*        style={{*/}
                {/*            width: '100%',*/}
                {/*            maxWidth: '800px',*/}
                {/*            margin: '2em auto',*/}
                {/*            height: '50px',*/}
                {/*            fontSize: '1.2em',*/}
                {/*            borderRadius: '5px',*/}
                {/*            border: '1px solid rgba(var(--foreground), 0.1)',*/}
                {/*            backgroundColor: 'rgba(var(--foreground), 0.01)',*/}
                {/*            backdropFilter: 'blur(10px)',*/}
                {/*            textAlign: 'center',*/}
                {/*        }}*/}
                {/*        onChange={(e) => {*/}
                {/*            console.log(e.target.value);*/}
                {/*        }}*/}
                {/*    />*/}
                {/*</div>*/}
                <div className="flex" style={{flexWrap: 'wrap'}}>
                    {tags.map((tag, index) => (
                        <TagItemCard name={tag.tagName} count={tag.articleCount} index={index} key={tag.tagId}/>
                    ))}
                </div>
            </Container>
            <TagCloudBackground tags={tags.map((tag) => tag.tagName)}/>
        </div>
    );
};

export default Page;
