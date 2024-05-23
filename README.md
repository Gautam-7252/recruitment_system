# Recruitment_system
<div><h2>signup</h2></div>
<div>curl -X POST http://localhost:8000/signup \
     -H "Content-Type: application/json" \
     -d '{"name": "John Doe", "email": "johndoe@example.com", "password": "your_password", "user_type": "Applicant", "profile_headline": "Software Engineer", "address": "123 Main St"}'</div><br>

<div><h2>Login</h2></div>
<div>curl -X POST http://localhost:8000/login \
     -H "Content-Type: application/json" \
     -d '{"email": "johndoe@example.com", "password": "your_password"}'</div><br>

<div><h2>Upload Resume</h2></div>
<div>curl -X POST http://localhost:8000/uploadResume \
     -H "Authorization: Bearer <your_jwt_token>" \
     -F "resume=@path/to/resume.pdf"</div><br>

<div><h2>Admin Create Job</h2></div>
<div>curl -X POST http://localhost:8000/admin/job \
     -H "Authorization: Bearer <your_jwt_token>" \
     -H "Content-Type: application/json" \
     -d '{"title": "Software Engineer", "description": "Job description", "company_name": "ABC Inc."}'</div><br>

<div><h2>Get job</h2></div>
<div>curl http://localhost:8000/admin/job/1 \
     -H "Authorization: Bearer <your_jwt_token>"</div><br>

<div><h2>Get jobs</h2></div>
<div>curl http://localhost:8000/jobs \
     -H "Authorization: Bearer <your_jwt_token>"</div><br>


<div><h2>Get applicants</h2></div>
<div>curl http://localhost:8000/admin/applicants \
     -H "Authorization: Bearer <your_jwt_token>"</div><br>

<div><h2>Get applicant</h2></div>
<div>curl http://localhost:8000/admin/applicant/1 \</div><br>
     -H "Authorization: Bearer <your_jwt_token>"

Apply jobs
curl http://localhost:8000/jobs/apply?job_id=1 \
     -H "Authorization: Bearer <your_jwt_token>"
