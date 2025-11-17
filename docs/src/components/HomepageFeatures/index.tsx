import clsx from "clsx";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";
import { JSX } from "react";

type FeatureItem = {
  title: string;
  icon: string;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: "Обход DPI в реальном времени",
    icon: "🛡️",
    description: (
      <>
        Продвинутая манипуляция пакетами с TCP фрагментацией, UDP маскировкой и
        SNI спуфингом
      </>
    ),
  },
  {
    title: "Веб-интерфейс управления",
    icon: "🎛️",
    description: (
      <>
        Красивая панель управления с метриками в реальном времени, потоковыми
        логами и управлением конфигурацией
      </>
    ),
  },
  {
    title: "Интеграция GeoIP/GeoSite",
    icon: "🌍",
    description: (
      <>
        Поддержка геоданных v2ray/xray с автоматическими обновлениями и
        фильтрацией по категориям
      </>
    ),
  },
  {
    title: "Мульти-сет конфигурации",
    icon: "⚙️",
    description: (
      <>
        Создавайте множественные стратегии обхода с различными параметрами для
        разных сценариев
      </>
    ),
  },
  {
    title: "Сетевая аналитика",
    icon: "🔍",
    description: <>Интеграция с IPInfo и RIPE для поиска ASN и анализа сетей</>,
  },
  {
    title: "Высокая производительность",
    icon: "⚡",
    description: (
      <>Многопоточная обработка пакетов с минимальным влиянием на задержку</>
    ),
  },
];

function Feature({ title, icon, description }: FeatureItem) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center">
        <div className="feature-icon">{icon}</div>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
