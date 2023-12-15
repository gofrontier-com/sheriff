.. image:: https://pkg.go.dev/badge/github.com/frontierdigital/sheriff.svg
    :target: https://pkg.go.dev/github.com/frontierdigital/sheriff
.. image:: https://github.com/frontierdigital/sheriff/actions/workflows/ci.yml/badge.svg
    :target: https://github.com/frontierdigital/sheriff/actions/workflows/ci.yml

=======
Sheriff
=======

Sheriff is a command line tool to manage Azure role-based access control (RBAC)
and Microsoft Entra Priviliged Identity Management (PIM) configuration declaratively.

.. contents:: Table of Contents
    :local:

-----
About
-----

Sheriff has been built to enable the management of Azure RBAC and Microsoft Entra PIM configuration
via YAML/JSON files. Although some of its functionality overlaps with the AzureRM provider
for Terraform, the Terraform implementation lacks coverage for some key features required
to operate PIM effectively, including role management policies.

Where Terraform also requires state to be maintained, Sheriff is different: it uses Azure APIs as it's
only source of truth, and ensures configuration is always consistent with the desired state, regardless
of how that configuration was set. For example, if a user manually adds a role assignment that isn't
present in the desired state YAML configuration, Sheriff will remove it.

Sheriff is designed to be used as part of a CI/CD pipeline.

--------
Download
--------

~~~~~~~
Release
~~~~~~~

Binaries and packages of the latest stable release are available at `https://github.com/frontierdigital/sheriff/releases <https://github.com/frontierdigital/sheriff/releases>`_.

~~~~~~~~~
Extension
~~~~~~~~~

The Sheriff extension for Azure DevOps is available from `Visual Studio Marketplace <https://marketplace.visualstudio.com/items?itemName=frontierdigital.sheriff>`_, which will automatically install Sheriff via a task.

-----
Usage
-----
