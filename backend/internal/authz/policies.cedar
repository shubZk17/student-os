// StudentOS authorization policy (Cedar).
//
// Every student-owned record carries an `owner` attribute. A student may act on a
// record only when they are that owner. Before this file existed the same rule was
// implied by `AND user_id = $n` repeated across nine handlers, where it could not be
// read in one place or tested on its own.
//
// This policy is checked BEFORE the database call. The SQL ownership predicate is
// deliberately kept as well: Cedar is the auditable statement of intent, the SQL
// predicate is the backstop that still holds if a handler forgets to ask Cedar.

// A student may read and modify their own application tracker entries.
permit (
    principal,
    action in [
        Action::"viewApplication",
        Action::"updateApplication",
        Action::"deleteApplication"
    ],
    resource
)
when { resource.owner == principal };

// A student may read and modify their own portfolio projects.
permit (
    principal,
    action in [
        Action::"viewProject",
        Action::"createProject",
        Action::"deleteProject"
    ],
    resource
)
when { resource.owner == principal };

// A student may read and mark their own notifications.
permit (
    principal,
    action in [
        Action::"viewNotification",
        Action::"markNotificationRead"
    ],
    resource
)
when { resource.owner == principal };

// A student may read and update their own profile.
permit (
    principal,
    action in [Action::"viewProfile", Action::"updateProfile"],
    resource
)
when { resource.owner == principal };

// Opportunities are public catalogue data: any authenticated student may read one.
// There is no ownership check here because postings belong to no student.
permit (
    principal,
    action == Action::"viewOpportunity",
    resource
);

// Everything not permitted above is denied, because Cedar is deny-by-default.
// There is intentionally no policy granting cross-student access, and no admin
// role: StudentOS has no administrative surface today.
