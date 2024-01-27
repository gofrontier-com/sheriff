.. image:: https://pkg.go.dev/badge/github.com/gofrontier-com/sheriff.svg
    :target: https://pkg.go.dev/github.com/gofrontier-com/sheriff
.. image:: https://github.com/gofrontier-com/sheriff/actions/workflows/ci.yml/badge.svg
    :target: https://github.com/gofrontier-com/sheriff/actions/workflows/ci.yml

|

.. image:: logo.png
  :width: 200
  :alt: Sheriff logo
  :align: center

=======
Sheriff
=======

Sheriff is a command line tool to manage **Azure role-based access control (Azure RBAC)**
and **Microsoft Entra ID Privileged Identity Management (Microsoft Entra ID PIM)** using desired state configuration.

.. contents:: Table of Contents
    :local:

-----
About
-----

~~~~~~~
Sheriff
~~~~~~~

Sheriff has been built to enable the management of Azure RBAC and Microsoft Entra ID PIM configuration
via YAML/JSON files. Although some of its functionality overlaps with the AzureRM provider
for Terraform, the Terraform implementation lacks coverage for some key features required
to operate PIM effectively, including role management policies.

Where Terraform also requires state to be maintained, Sheriff is different: it uses Azure APIs as it's
only source of truth, and ensures configuration is always consistent with the desired state, regardless
of how that configuration was set. For example, if a user manually adds a role assignment that isn't
present in the desired state YAML configuration, Sheriff will remove it.

Sheriff is designed to be used as part of a CI/CD pipeline.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Azure role-based access control (Azure RBAC)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Azure role-based access control (Azure RBAC) helps you manage who has access to
Azure resources, what they can do with those resources, and what areas they have access to.

See `What is Azure role-based access control (Azure RBAC)? <https://learn.microsoft.com/en-us/azure/role-based-access-control/overview>`_ for more information.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Microsoft Entra ID Privileged Identity Management (Microsoft Entra ID PIM)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Microsoft Entra ID Privileged Identity Management (PIM) is a service in Microsoft Entra ID that
enables you to manage, control, and monitor access to important resources in your organization.
These resources include resources in Microsoft Entra ID, Azure, and other Microsoft Online Services
such as Microsoft 365 or Microsoft Intune.

See `What is Microsoft Entra ID Privileged Identity Management? <https://learn.microsoft.com/en-gb/entra/id-governance/privileged-identity-management/pim-configure?WT.mc_id=Portal-Microsoft_Azure_PIMCommon>`_ for more information.

--------
Download
--------

~~~~~~~
Release
~~~~~~~

Binaries and packages of the latest stable release are available at `https://github.com/gofrontier-com/sheriff/releases <https://github.com/gofrontier-com/sheriff/releases>`_.

~~~~~~~~~
Extension
~~~~~~~~~

The Sheriff extension for Azure DevOps is available from `Visual Studio Marketplace <https://marketplace.visualstudio.com/items?itemName=gofrontier.sheriff>`_, which will automatically install Sheriff via a task.

-------------
Configuration
-------------

~~~~~~~~~~~~~~~
Azure Resources
~~~~~~~~~~~~~~~

.. code:: bash

  groups/
    <group name>.yml
    ...
  users/
    <user upn>.yml
    ...
  policies/
    <role name>.yml
    rulesets/
      <ruleset name>.yml
    ...

Configuration of active and eligible role assigments is managed via YAML files per group and/or user,
in which both active and eligible role assignments are defined.

``groups/<group name>.yml`` or ``users/<user upn>.yml``

.. code:: yaml

  ---
  subscription:
    active:
      - roleName: <role name>
      ...
    eligible:
      - roleName: <role name>
      ...
  resourceGroups:
    <resource group name>:
      active:
        - roleName: <role name>
        ...
      eligible:
        - roleName: <role name>
        ...
  resources:
    <resource name>:
      active:
        - roleName: <role name>
        ...
      eligible:
        - roleName: <role name>
        ...

Configuration of role management policies is managed via YAML files per role.
Role configuration files reference one or more rulesets at the required scopes.

!!!Important!!! Please note that role management policies are **not** inherited from parent scopes.
This is by design in Microsoft Entra ID PIM and cannot be changed. Overriding the
default role management policy for a given role at a particular scope must be done
by referencing one or more rulesets at that exact scope.

``policies/<role name>.yml``

.. code:: yaml

  ---
  subscription:
    - rulesetName: <ruleset name>
    ...
  resourceGroups:
    <resource group name>:
      - rulesetName: <ruleset name>
      ...
  resources:
    <resource name>:
      - rulesetName: <ruleset name>
      ...

Rules (and partial rules) defined in rulesets override those in the
`default role management policy <https://github.com/gofrontier-com/azurerm-terraform-modules/tree/main/pkg/cmd/app/apply/default_role_management_policy.json>`_.

``policies/rulesets/<ruleset name>.yml``

.. code:: yaml

  ---
  rules:
    - id: Approval_EndUser_Assignment
      patch:
        setting:
          approvalStages:
            - approvalStageTimeOutInDays: 1
              escalationTimeInMinutes: 0
              isApproverJustificationRequired: true
              isEscalationEnabled: false
              primaryApprovers:
                - userType: Group
                  isBackup: false
                  id: abd8337a-b700-4de5-a800-006d893fc015
                  description: CSG-RBAC-SeniorEngineers
          isApprovalRequired: true

See `Rules in PIM - mapping guide <https://learn.microsoft.com/en-us/graph/identity-governance-pim-rules-overview>`_ for more information.

Examples
~~~~~~~~

Active assignment for group at subscription scope
-------------------------------------------------

``groups/Engineers.yml``

.. code:: yaml

  ---
  subscription:
    active:
      - roleName: Reader

Active assignment for user at resource group scope
--------------------------------------------------

``users/john@gofrontier.com.yml``

.. code:: yaml

  ---
  resourceGroups:
    rg-dev-virtualmachine:
      active:
        - roleName: Contributor

Active assignment for user at resource scope
--------------------------------------------

``users/john@gofrontier.com.yml``

.. code:: yaml

  ---
  resources:
    rg-dev-virtualnetwork/providers/Microsoft.Network/virtualNetworks/vnet-dev-main:
      active:
        - roleName: Network Contributor

Eligible assignment for group at subscription scope
---------------------------------------------------

``groups/SRE.yml``

.. code:: yaml

  ---
  subscription:
    eligible:
      - roleName: Disk Restore Operator
        endDateTime: 2024-12-31T00:00:00Z

By default, Entra ID PIM requires that eligible assignments have an expiry date. To create an eligible assignment that never expires, you must create a role management policy ruleset that disables this requirement.

``policies/Disk Restore Operator.yml``

.. code:: yaml

  ---
  subscription:
    - rulesetName: NoEligibleExpiry

``policies/rulesets/NoEligibleExpiry.yml``

.. code:: yaml

  ---
  rules:
    - id: Expiration_Admin_Eligibility
      patch:
        isExpirationRequired: false

With the above created, you can now omit an expiry date.

``groups/SRE.yml``

.. code:: yaml

  ---
  subscription:
    eligible:
      - roleName: Disk Restore Operator

Eligible assignment for user at resource scope with approval
------------------------------------------------------------

``policies/rulesets/ApprovalRequired.yml``

.. code:: yaml

  ---
  rules:
    - id: Approval_EndUser_Assignment
      patch:
        setting:
          approvalStages:
            - approvalStageTimeOutInDays: 1
              escalationTimeInMinutes: 0
              isApproverJustificationRequired: true
              isEscalationEnabled: false
              primaryApprovers:
                - userType: Group
                  isBackup: false
                  id: abd8337a-b700-4de5-a800-006d893fc015
                  description: SeniorEngineers
          isApprovalRequired: true

``policies/Network Contributor.yml``

.. code:: yaml

  ---
  resources:
    rg-dev-virtualnetwork/providers/Microsoft.Network/virtualNetworks/vnet-dev-main:
      - rulesetName: ApprovalRequired
      - rulesetName: NoEligibleExpiry

``users/john@gofrontier.com.yml``

.. code:: yaml

  ---
  resources:
    rg-dev-virtualnetwork/providers/Microsoft.Network/virtualNetworks/vnet-dev-main:
      eligible:
        - roleName: Network Contributor

~~~~~~~~~~~~~~~~~~~~~
Microsoft Entra roles
~~~~~~~~~~~~~~~~~~~~~

*Coming soon...*

~~~~~~
Groups
~~~~~~

*Coming soon...*

-----
Usage
-----

.. code:: bash

  $ sheriff --help
  Sheriff is a command line tool to manage Azure role-based access control (RBAC) and Microsoft Entra ID Privileged Identity Management (PIM) configuration declaratively

  Usage:
    sheriff
    sheriff [command]

  Available Commands:
    apply       Apply config
    completion  Generate the autocompletion script for the specified shell
    help        Help about any command
    plan        Plan changes
    validate    Validate config
    version     Output version information

  Flags:
    -h, --help   help for sheriff

  Use "sheriff [command] --help" for more information about a command.

~~~~~~~~~~~~~~~
Azure Resources
~~~~~~~~~~~~~~~

Plan
~~~~

.. code:: bash

  $ sheriff plan azurerm \
      --config-dir <path to AzureRM config> \
      --subscription-id <subscription ID>

Apply
~~~~~

.. code:: bash

  $ sheriff apply azurerm \
      --config-dir <path to AzureRM config> \
      --subscription-id <subscription ID>

~~~~~~~~~~~~~~~~~~~~~
Microsoft Entra roles
~~~~~~~~~~~~~~~~~~~~~

*Coming soon...*

~~~~~~
Groups
~~~~~~

*Coming soon...*

------------
Contributing
------------

We welcome contributions to this repository. Please see `CONTRIBUTING.md <https://github.com/gofrontier-com/azurerm-terraform-modules/tree/main/CONTRIBUTING.md>`_ for more information.
