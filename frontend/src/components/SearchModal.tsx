"use client"

import * as React from "react"
import {Button, Dialog, Flex, Text, ScrollArea} from "@radix-ui/themes"
import {MagnifyingGlassIcon, Cross2Icon} from "@radix-ui/react-icons"
import {Command} from "cmdk"
import "@radix-ui/themes/styles.css"
import {useRouter} from "next/navigation"
import {debounce} from "lodash"
import {searchItems} from "@/api/search"
import {SearchResults} from "@/components/SearchResults";

export default function SearchModal({
                                        open,
                                        setOpen,
                                    }: {
    open: boolean
    setOpen: React.Dispatch<React.SetStateAction<boolean>>
}) {
    const [query, setQuery] = React.useState("")
    const [searchResults, setSearchResults] = React.useState(null)
    const router = useRouter()

    const handleSearch = React.useMemo(
        () =>
            debounce(async (searchTerm: string) => {
                if (searchTerm.length > 0) {
                    const results = await searchItems(searchTerm)
                    setSearchResults(results)
                } else {
                    setSearchResults(null)
                }
            }, 300),
        []
    )

    React.useEffect(() => {
        handleSearch(query)
        return () => handleSearch.cancel()
    }, [handleSearch, query])

    React.useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
                e.preventDefault()
                setOpen((open) => !open)
            }
        }
        document.addEventListener("keydown", down)
        return () => document.removeEventListener("keydown", down)
    }, [setOpen])

    return (
        <Dialog.Root open={open} onOpenChange={setOpen}>
            <Dialog.Content
                style={{
                    maxWidth: "750px",
                    padding: "0",
                    backgroundColor: "rgba(var(--background), 0.8)",
                    backdropFilter: "blur(10px)",
                    borderRadius: "12px",
                    boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.25)",
                    overflow: "hidden",
                }}
            >
                <Command
                    style={{
                        display: "flex",
                        flexDirection: "column",
                        height: "450px",
                    }}
                >
                    <Flex align="center" style={{padding: "16px", borderBottom: "1px solid var(--gray-a5)"}}>
                        <MagnifyingGlassIcon width="20" height="20"
                                             style={{marginRight: "12px", color: "var(--gray-a11)"}}/>
                        <Command.Input
                            autoFocus
                            placeholder="搜索任何内容..."
                            value={query}
                            onValueChange={(value) => setQuery(value)}
                            style={{
                                width: "100%",
                                border: "none",
                                background: "transparent",
                                fontSize: "16px",
                                outline: "none",
                                color: "var(--gray-12)",
                            }}
                        />
                        <Dialog.Close>
                            <Button variant="ghost" color="gray" style={{padding: "8px", borderRadius: "6px"}}>
                                <Cross2Icon width="18" height="18"/>
                            </Button>
                        </Dialog.Close>
                    </Flex>
                    {searchResults ? (
                        <SearchResults results={searchResults}/>
                    ) : (
                        <ScrollArea style={{flex: 1, padding: "8px"}}>
                            <Command.List className="pl-4 pr-4">
                                <Command.Group heading="建议">
                                    <CommandItem onClick={() => router.push("/posts")}> 所有文章 </CommandItem>
                                    <CommandItem onClick={() => router.push("/moments")}> 所有随手记录 </CommandItem>
                                    <CommandItem onClick={() => router.push("/archives")}> 所有归档 </CommandItem>
                                </Command.Group>
                                <Command.Group heading="设置">
                                    <CommandItem onClick={() => router.push("/my")}> 个人资料 </CommandItem>
                                </Command.Group>
                            </Command.List>
                        </ScrollArea>
                    )}
                    <Flex
                        justify="between"
                        align="center"
                        style={{
                            borderTop: "1px solid var(--gray-a5)",
                            padding: "12px 16px",
                            backgroundColor: "var(--gray-a2)",
                        }}
                    >
                        <Text size="1" style={{color: "var(--gray-a11)"}}>
                            搜索由 MeiliSearch 强力驱动
                        </Text>
                        <Flex gap="2">
                            <kbd
                                style={{
                                    padding: "2px 6px",
                                    borderRadius: "4px",
                                    backgroundColor: "var(--gray-a4)",
                                    color: "var(--gray-a11)",
                                    fontSize: "12px",
                                }}
                            >
                                ↑↓
                            </kbd>
                            <Text size="1" style={{color: "var(--gray-a11)"}}>
                                导航
                            </Text>
                            <kbd
                                style={{
                                    padding: "2px 6px",
                                    borderRadius: "4px",
                                    backgroundColor: "var(--gray-a4)",
                                    color: "var(--gray-a11)",
                                    fontSize: "12px",
                                }}
                            >
                                ↵
                            </kbd>
                            <Text size="1" style={{color: "var(--gray-a11)"}}>
                                选择
                            </Text>
                        </Flex>
                    </Flex>
                </Command>
            </Dialog.Content>
        </Dialog.Root>
    )
}

function CommandItem({children, onClick}: { children: React.ReactNode; onClick?: () => void }) {
    return (
        <Command.Item
            onSelect={onClick}
            style={{
                display: "flex",
                alignItems: "center",
                padding: "8px 12px",
                borderRadius: "6px",
                cursor: "pointer",
                transition: "all 0.2s ease-in-out",
                color: "var(--gray-12)",
            }}
            onMouseEnter={(e) => {
                e.currentTarget.style.backgroundColor = "var(--accent-a5)"
            }}
            onMouseLeave={(e) => {
                e.currentTarget.style.backgroundColor = "transparent"
            }}
        >
            {children}
        </Command.Item>
    )
}
