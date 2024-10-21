**Zoha** is a lightweight **LMTP** (Local Mail Transfer Protocol) server written in Go (golang) designed for use with various **Mail Transfer Agents (MTA)**, most commonly paired with **Courier IMAP** and **Postfix**. This combination allows for the creation of a highly scalable, distributed email system. Although Zoha was specifically developed to integrate seamlessly with Courier IMAP and Postfix, it remains flexible and can operate with other MTAs as well.

In such a setup:

* **Postfix** manages email routing, ensuring messages are delivered between servers.
* **Courier IMAP** provides storage and retrieval of emails, allowing users to access their inboxes via IMAP.
* **Zoha** facilitates efficient local email delivery to storage using LMTP, which is optimized for handling high volumes of email in distributed environments.
 
This combination is particularly suited for systems requiring scalability across multiple servers, offering both performance and reliability.

### Where did the idea to create Zoha come from?

The idea behind Zoha was inspired by studying the hardware-software architecture used by companies like Google. Google's infrastructure relies on a large number of relatively inexpensive machines, which can be quickly and easily replaced when needed. This approach contrasts with the traditional use of high-performance, expensive hardware.

In email systems, where intensive I/O operations are a significant challenge, maintaining high performance typically requires durable and fast storage solutions, such as modern NVMe drives. However, the Zoha architecture offers an alternative by distributing the workload across multiple average-performing nodes. This model provides increased fault tolerance because individual node failures only affect a small subset of clients. In such a setup, downtime is minimized, and reliability is enhanced, as nodes can be easily replaced or scaled up without significant disruptions.

This distributed, commodity-based infrastructure allows hosting providers to handle large volumes of email traffic efficiently, reducing the dependency on expensive hardware while improving system resilience. The idea is that instead of relying on a few high-performance servers, you can achieve greater reliability and scalability with many less costly, replaceable machines.

### Zoha performance

The system built on Zoha has been running in a production environment for nearly 12 months, successfully managing tens of thousands of business accounts. This demonstrates Zoha's reliability and scalability in handling large-scale email operations.

### Where does the name come from?

*"Zocha"* (eng. Sophia) is a somewhat informal, even slightly cheeky, diminutive, typically used to describe someone with strong, resilient traits. It implies toughness and endurance, characteristics attributed to the fantastic fox terrier that inspired the name. This particular dog, despite facing illness and pain, consistently demonstrated an ability to endure hardship without showing signs of weakness, embodying the qualities of strength, perseverance, and an indomitable spirit.

Zoha, as a software solution, reflects these qualities—resilience in the face of challenges, reliability under stress, and a small but powerful presence, perfectly suited for high-demand environments like scalable email systems. This metaphor encapsulates the project's philosophy: that great strength and capability can come from modest beginnings. The system, much like the terrier, is built to endure and perform even in difficult circumstances, making it a symbol of fortitude and robustness.

Marcin Maluszczak