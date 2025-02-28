FROM locustio/locust:latest

# Install dependencies
RUN pip install --no-cache-dir transformers numpy

# Copy your locust file
COPY llm_user.py /home/locust/locustfile.py

# The default entrypoint from locustio/locust image is:
ENTRYPOINT ["locust"]