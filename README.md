# Stratometrics

Stratometrics is an open-source billing metrics collection system, meticulously
crafted to provide granular, event-based billing data for OpenStack cloud
environments.

As a pre-installed component within [Atmosphere](https://github.com/vexxhost/atmosphere),
an OpenStack distribution powered by fully open source components, it offers a
seamless billing metrics experience out of the box.

## Overview

This system is engineered to capture detailed resource usage events with
sub-millisecond precision, empowering Atmosphere users with the ability to track
resource consumption accurately for billing purposes. Administrators gain access
to a wealth of data, enabling sophisticated billing models and ensuring fair,
transparent cost allocation.

Utilizing Golang's performance capabilities, Stratometrics delivers
high-throughput data collection, while leveraging Clickhouse for its backend
database to ensure quick processing and retrieval of billing-related data. The
API provides easy and direct access to usage metrics, supporting the generation
of comprehensive billing statements and analytics.

## Features

- **Sub-Millisecond Billing Accuracy**: Captures billing events with the finest
  granularity for unparalleled precision.
- **Integrates with OpenStack**: Seamlessly integrated within the Atmosphere
  OpenStack distribution for immediate use, but works with any OpenStack deployment.
- **User & administrative API**: Equipped with APIs tailored for both
  user-specific and administrative billing inquiries.
- **Robust Data Handling**: Backed by Clickhouse, it manages large-scale billing
  data with efficiency.

## Getting Started

TODO.

## TODO

- Additional billed resources (floating IPs, etc.)

## Contributing

Your contributions make Stratometrics better. For enhancements, fixes, or other
contributions, please submit pull requests or engage in discussions via the
issues tracker.

## Support

This is a community supported project.  If you require commercial support, please
contact [VEXXHOST](https://vexxhost.com).
