.. _multi_user_accounts:

Multi-user Accounts
###################

Multi-user accounts help teams share services and infrastructure on |ITAC|.
This feature simplifies collaboration, billing, and infrastructure management.

.. note::
   Multi-user accounts are only available to premium and enterprise account members.

Prerequisites
**************

* :ref:`get_started`
* :ref:`accounts`

Roles
======

An **administrator**:

* *creates* the premium or enterprise account
* may invite several users to join multi-user accounts
* chooses an expiration date for an invitation

A **member** must:

* accept an invitation to become a member
* join an account before an invitation expires

Follow these instructions based on your `Administrator`_ or `Member`_ role.

Administrator
*************

Inviting a new member into a premium or enterprise account is a two-step process for the account administrator: Send an invitation and verify your administrator role.

Send invitation
===============

#. Sign into the console.

#. Click the :guilabel:`User Profile` pull-down menu in upper right.

#. In the drop-down menu, click :guilabel:`Account Settings`.

#. View :guilabel:`Account Access Management` at bottom.

#. Click :guilabel:`Grant access` to invite individuals.

   In the dialog, enter:

   * **Email** - Email address of the person you want to join

   * **Invitation Expiration** - Expiration date of the invitation

   * **Note** - Optional: Add a note to be sent with the invitation

#. Click :guilabel:`Grant`.


Verify
======

After sending an invitation, a verification code is emailed to the administrator's email address for security.
An administrator uses that code to verify this is a legitimate request to add a user to the account.

#. Open the administrator verification email in your email account and copy the verification code.

#. Return to the console and find the dialog menu where you enter the verification code.

#. Paste the verification code into the empty field in the dialog box.

#. Click :guilabel:`Verify`.

#. In :guilabel:`Account Access Management`, identify the user you invited.

#. Review the :guilabel:`Status` column to assure it reflects your desired expiration date.

#. View the options under the column :guilabel:`Invitation Actions`.

#. While an invitation is in a *pending state*, you may:

   * Remove
   * Resend

   .. tip::
      *After a user accepts an invitation*, you may select :guilabel:`Revoke` to revoke membership for the user
      you invited to the multi-user account.

#. Remind users to check their email accounts for the invitation.

Revoke Access
=============

As the Administrator, you may revoke access for a member's SSH keys at any time.

To revoke access, follow **all steps** below.

Revoke Access for Member
------------------------

#. Access the user profile icon.

#. Select :guilabel:`Account Settings`.

#. Select :guilabel:`Members`

#. In :guilabel:`Account Access Management`, find the :guilabel:`Invitation` column.

#. Select the user for which you wish to revoke access.

#. Select :guilabel:`Revoke`.

#. Click :guilabel:`Revoke Access` and confirm.

#. Continue below.

Connect to your Instance
------------------------

#. Navigate to :guilabel:`Compute -> Instances`.

#. Click your instance under :guilabel:`Instance Name`.

#. Click :guilabel:`How to Connect via SSH`.

#. Copy the "SSH command" to connect to instance.

#. Open a Terminal to connect to your instance.

#. Paste the "SSH command" from a previous step.

#. Log into your instance.

Determine Member SSH Key to Delete
----------------------------------

#. In the console app, identify the member whose SSH Keys you wish to remove.

   #. Navigate to :guilabel:`Compute -> Keys`.

   #. In the :guilabel:`Keys` tab, find the member under :guilabel:`Name`.

   #. Click :guilabel:`Copy Key`.

   #. Paste the **value** of that member SSH key in a blank document for reference.

   #. Continue below.

Delete SSH Keys of Member
-------------------------

#. With your preferred editor, open the file :file:`~/.ssh/authorized_keys`.

   .. tip::
      See **Optional** section below to install an editor.

#. Locate the member's SSH Key to be deleted. Compare the member key found in previous section with the one that appears in the file.

   .. warning::
      Be very careful editing :file:`authorized_keys`. If your main SSH key is changed or deleted, you will lose SSH access to the instance.
      See also :ref:`ssh_keys`.

#. Delete the line containing the member's SSH Keys.

#. Save the file.

Install Editor on Linux* OS
---------------------------

#. **Optional**: To install an editor on your instance:

   #. Run :command:`sudo apt update`

   #. Run :command:`sudo apt install vim`

      Alternatively, replace ``vim`` with ``nano`` in previous command.

Remove Member Key from Instance
---------------------------------

#. Navigate to :guilabel:`Compute -> Instances`.

#. In the row of the named instance, select :guilabel:`Edit`.

#. In :guilabel:`Edit instance`, under :guilabel:`Public Keys`, **uncheck** the member key you wish to delete.

#. Now select your own key.  You must select at least one key.

#. Click :guilabel:`Save`.

#. In the dialog, select :guilabel:`Close`.


Member
******

Follow these steps to *become a member* when an administrator sends you an invitation.

Join account
============

You\'ll receive an email invitation to join a multi-user account.
Follow the instructions in the email. Then continue below.

#. Sign into the console.

#. A dialog menu appears, requesting you to :guilabel:`Accept` or :guilabel:`Decline` an invitation
   to a multi-user account.

   .. note::
      Accepting this invitation enables you to use the resources of the multi-user
      account. However, restrictions apply to account management.

#. Click :guilabel:`Accept` to accept the invitation.

#. Check for a verification email and copy the one-time password (OTP).

#. In the console, find the invitation code text field where you'll enter the OTP.

#. Paste the OTP into the text field.

#. Click :guilabel:`Confirm`.

#. Verify that the notification, :guilabel:`Invitation confirmed`, appears.

You\'ve successfully joined a multi-user account.


Switch accounts
***************

Use the :guilabel:`Switch Accounts` feature to toggle between multi-user accounts and your own account.

#. In the user profile, select the pull down menu and click :guilabel:`Switch Accounts`.

   .. note::
      If you\'re already logged in, use the :guilabel:`Switch Accounts` menu (at top).

#. If you\'re just logging into your account, a dialog may appear requesting that you:

   * Select an account; or

   * :guilabel:`Accept` or :guilabel:`Decline` an invitation.

#. If you accept the invitation, you can start using the shared account at the Premium/Enterprise tier level
   (regardless of whether your individual account is standard).

.. tip::
   You may have multiple accounts, each of which may have a different tier level.


FAQ
***

.. list-table::
   :header-rows: 1
   :class: table-tiber-theme

   * - Question
     - Answer

   * - Is a multi-user account allowed between two or more Standard tier account holders?
     - No. However, if a Premium/Enterprise user invites a Standard tier user to join an account,
       such user may utilize the Premium/Enterprise services and features.

   * - Can users who receive the invitation remove themselves from the multi-user account?
     - No. Only the cloud account owner is allowed to invite/remove members.

   * - Can a user become a member of more than one multi-user accounts?
     - Yes. Absolutely.

   * - How do I switch between accounts?
     - Use :guilabel:`Switch Accounts` to change the active account. Note that
       this menu only appears after you\'ve accepted an invitation.

