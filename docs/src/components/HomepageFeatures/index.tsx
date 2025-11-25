import clsx from "clsx";
import Heading from "@theme/Heading";
import Link from "@docusaurus/Link";
import styles from "./styles.module.css";
import { JSX } from "react";

type DocSection = {
  title: string;
  link: string;
  description: JSX.Element;
};

const DocSections: DocSection[] = [
  {
    title: "Установка",
    link: "/docs/intro",
    description: (
      <>
        Быстрая установка на Linux, OpenWRT, Entware. Настройка systemd, запуск
        сервиса, решение типичных проблем.
      </>
    ),
  },
  {
    title: "Мониторинг соединений",
    link: "/docs/domains",
    description: (
      <>
        Live-монитор TCP/UDP трафика. Добавление доменов и IP в обход DPI,
        обогащение через RIPE/IPInfo, фильтрация потока.
      </>
    ),
  },
  {
    title: "Сеты конфигураций",
    link: "/docs/sets",
    description: (
      <>
        Создание наборов настроек для разных сценариев. TCP/UDP параметры,
        стратегии фрагментации, фейкинг, мутация ClientHello.
      </>
    ),
  },
  {
    title: "Общие настройки",
    link: "/docs/settings",
    description: (
      <>
        Сетевые параметры, логирование, firewall, geodata файлы, внешние API,
        захват пакетов.
      </>
    ),
  },
];

function DocCard({ title, link, description }: DocSection) {
  return (
    <div className={clsx("col col--6")}>
      <div className={styles.docCard}>
        <Heading as="h3">
          <Link to={link}>{title}</Link>
        </Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <Heading as="h2" className={styles.sectionTitle}>
          Документация
        </Heading>
        <div className="row">
          {DocSections.map((props, idx) => (
            <DocCard key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
