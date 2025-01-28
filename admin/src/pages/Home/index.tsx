import type React from "react"
import { useEffect, useState } from "react"
import { PageContainer } from "@ant-design/pro-components"
import { Card, Typography, Row, Col, Statistic } from "antd"
import type { OneWordResponse } from "@/services/oneword/typings"
import type { Overview } from "@/services/overview/typings"
import OneWordController from "@/services/oneword/OneWordController"
import OverviewController from "@/services/overview/OverviewController"
import styles from "./HomePage.module.css"

const { Title, Paragraph } = Typography

const HomePage: React.FC = () => {
  const [oneWord, setOneWord] = useState<OneWordResponse>()
  const [overview, setOverview] = useState<Overview>()

  useEffect(() => {
    OneWordController.refreshOneWord().then((res) => {
      setOneWord(res)
    })
    OverviewController.getOverview().then((res) => {
      setOverview(res.data)
    })
  }, [])

  return (
      <PageContainer ghost>
        <div className={styles.container}>
          <Card className={styles.oneWordCard}>
            <Title level={2} className={styles.oneWordTitle}>
              每日一言
            </Title>
            <Paragraph className={styles.oneWord}>{oneWord?.hitokoto}</Paragraph>
            <Paragraph className={styles.oneWordFrom}>—— {oneWord?.from}</Paragraph>
            <Paragraph className={styles.oneWordCreator}>
              {oneWord?.creator} · {oneWord?.created_at}
            </Paragraph>
          </Card>

          <Title level={2} className={styles.overviewTitle}>
            网站概览
          </Title>
          <Row gutter={[16, 16]} className={styles.overviewContainer}>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Card hoverable className={styles.overviewCard}>
                <Statistic title="文章总数" value={overview?.articleCount} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Card hoverable className={styles.overviewCard}>
                <Statistic title="记录总数" value={overview?.momentCount} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Card hoverable className={styles.overviewCard}>
                <Statistic title="页面总数" value={overview?.pageCount} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Card hoverable className={styles.overviewCard}>
                <Statistic title="评论总数" value={overview?.commentCount} />
              </Card>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Card hoverable className={styles.overviewCard}>
                <Statistic title="阅读总数" value={overview?.allReadingCount} />
              </Card>
            </Col>
          </Row>
        </div>
      </PageContainer>
  )
}

export default HomePage

